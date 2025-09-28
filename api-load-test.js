import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics untuk rate limiting analysis
const errorRate = new Rate('errors');
const rateLimitRate = new Rate('rate_limit_errors');
const loginTrend = new Trend('login_duration');
const registerTrend = new Trend('register_duration');
const authRequestsCounter = new Counter('auth_requests');
const readRequestsCounter = new Counter('read_requests');
const writeRequestsCounter = new Counter('write_requests');
const rateLimitCounter = new Counter('rate_limit_hits');

// Configuration disesuaikan dengan rate limiting
// Auth: 10/min, Read: 120/min, Write: 60/min
export const options = {
    scenarios: {
        // Smoke test dengan rate limit friendly
        smoke_test: {
            executor: 'constant-vus',
            vus: 1,
            duration: '1m',
            tags: { test_type: 'smoke' },
        },

        // Load test dengan respect rate limiting
        // Target: Max 8 concurrent users untuk tidak exceed rate limits
        load_test: {
            executor: 'ramping-vus',
            startVUs: 2,
            stages: [
                { duration: '2m', target: 5 },   // Slow ramp up
                { duration: '5m', target: 8 },   // Conservative load
                { duration: '3m', target: 12 },  // Test near rate limits
                { duration: '2m', target: 5 },   // Back to safe zone
                { duration: '2m', target: 0 },   // Ramp down
            ],
            tags: { test_type: 'load' },
            startTime: '1m30s',
        },

        // Rate limit stress test - intentionally hit limits
        rate_limit_test: {
            executor: 'constant-arrival-rate',
            rate: 20, // 20 requests per second = 1200/min (exceed all limits)
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 5,
            maxVUs: 10,
            tags: { test_type: 'rate_limit_stress' },
            startTime: '16m',
        },
    },

    // Updated thresholds considering rate limiting
    thresholds: {
        http_req_duration: ['p(95)<2000'],        // More relaxed due to rate limiting
        http_req_failed: ['rate<0.3'],            // Accept higher failure rate due to rate limits
        rate_limit_errors: ['rate<0.5'],          // Track rate limit specifically
        login_duration: ['p(95)<800'],            // Account for potential delays
        register_duration: ['p(95)<1000'],
    },
};

const BASE_URL = 'http://localhost:8080/api';

// Helper functions
function generateRandomPhone() {
    const randomNum = Math.floor(Math.random() * 900000000) + 100000000;
    return `81${randomNum}`;
}

function generateRandomName() {
    const names = ['Alice', 'Bob', 'Charlie', 'Diana', 'Eva', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack'];
    return names[Math.floor(Math.random() * names.length)] + Math.floor(Math.random() * 1000);
}

function generateAccountNumber() {
    return Math.floor(Math.random() * 9000000000) + 1000000000;
}

function generateAccountName() {
    const firstNames = ['Ahmad', 'Budi', 'Citra', 'Dewi', 'Eko'];
    const lastNames = ['Santoso', 'Wijaya', 'Putri', 'Sari', 'Pratama'];
    return `${firstNames[Math.floor(Math.random() * firstNames.length)]} ${lastNames[Math.floor(Math.random() * lastNames.length)]}`;
}

// Check if response is rate limited
function isRateLimited(response) {
    return response.status === 429 ||
        (response.body && response.body.includes('Too many requests'));
}

// Setup function dengan retry mechanism untuk rate limiting
export function setup() {
    let attempts = 0;
    let maxAttempts = 3;

    while (attempts < maxAttempts) {
        const loginPayload = {
            number: '8123456789',
            password: '123456'
        };

        const loginRes = http.post(`${BASE_URL}/login`, JSON.stringify(loginPayload), {
            headers: { 'Content-Type': 'application/json' },
        });

        if (isRateLimited(loginRes)) {
            console.log(`⚠️  Rate limited on setup attempt ${attempts + 1}, waiting...`);
            sleep(15); // Wait before retry
            attempts++;
            continue;
        }

        const loginSuccess = check(loginRes, {
            'Setup Login status 200': (r) => r.status === 200,
            'Setup Login has access_token': (r) => r.json('data.access_token') !== undefined,
        });

        if (loginSuccess && loginRes.json('data.access_token')) {
            const token = loginRes.json('data.access_token');
            console.log('✅ Token obtained for authenticated requests');
            return { token: token };
        }

        attempts++;
        if (attempts < maxAttempts) sleep(10);
    }

    console.log('❌ Failed to get token after retries, authenticated requests will fail');
    return { token: null };
}

export default function (data) {
    const phone = generateRandomPhone();
    const name = generateRandomName();
    const token = data.token;

    // 1. Register new user (hanya 5% untuk menghormati rate limit auth: 10/min)
    if (Math.random() < 0.05) {
        const registerPayload = {
            name: name,
            number: phone,
            password: '123456',
            password_confirmation: '123456',
            referral_code: 'VLAREFF'
        };

        authRequestsCounter.add(1);
        const registerRes = http.post(`${BASE_URL}/register`, JSON.stringify(registerPayload), {
            headers: { 'Content-Type': 'application/json' },
        });

        const isRateLimit = isRateLimited(registerRes);
        if (isRateLimit) {
            rateLimitCounter.add(1);
            rateLimitRate.add(1);
        }

        const registerSuccess = check(registerRes, {
            'Register status 200/201 or rate limited': (r) => [200, 201, 429].includes(r.status) || isRateLimited(r),
            'Register has access_token (if success)': (r) => isRateLimited(r) || r.json('data.access_token') !== undefined,
            'Register response time reasonable': (r) => r.timings.duration < 2000,
        });

        registerTrend.add(registerRes.timings.duration);
        errorRate.add(!registerSuccess);

        if (isRateLimit) {
            console.log('📊 Register rate limited - expected behavior');
            sleep(Math.random() * 5 + 2); // Back off when rate limited
        }
    }

    // 2. Test Login endpoint (hanya 3% untuk menghormati rate limit auth: 10/min)
    if (Math.random() < 0.03) {
        const loginPayload = {
            number: '8123456789',
            password: '123456'
        };

        authRequestsCounter.add(1);
        const loginRes = http.post(`${BASE_URL}/login`, JSON.stringify(loginPayload), {
            headers: { 'Content-Type': 'application/json' },
        });

        const isRateLimit = isRateLimited(loginRes);
        if (isRateLimit) {
            rateLimitCounter.add(1);
            rateLimitRate.add(1);
        }

        const loginSuccess = check(loginRes, {
            'Login status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
            'Login has access_token (if success)': (r) => isRateLimited(r) || r.json('data.access_token') !== undefined,
            'Login response time reasonable': (r) => r.timings.duration < 1000,
        });

        loginTrend.add(loginRes.timings.duration);
        errorRate.add(!loginSuccess);

        if (isRateLimit) {
            console.log('📊 Login rate limited - expected behavior');
            sleep(Math.random() * 5 + 2);
        }
    }

    if (!token) {
        console.log('No token available, skipping authenticated requests');
        return;
    }

    const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
    };

    // Add longer delays to respect rate limits
    sleep(Math.random() * 2 + 1);

    // READ OPERATIONS (Rate limit: 120/min = 2/second)
    // 3. Get Products (70% of users - reduced from 90%)
    if (Math.random() < 0.7) {
        readRequestsCounter.add(1);
        const productsRes = http.get(`${BASE_URL}/products`, { headers });

        const isRateLimit = isRateLimited(productsRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(productsRes, {
            'Products status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
            'Products response time reasonable': (r) => r.timings.duration < 500,
        });

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 1 + 0.5);

    // 4. Get User Info (60% of users - reduced from 80%)
    if (Math.random() < 0.6) {
        readRequestsCounter.add(1);
        const userInfoRes = http.get(`${BASE_URL}/users/info`, { headers });

        const isRateLimit = isRateLimited(userInfoRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(userInfoRes, {
            'User info status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
            'User info response time reasonable': (r) => r.timings.duration < 400,
        });

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 1 + 0.5);

    // 5. Get Team Invited (40% of users - reduced from 60%)
    if (Math.random() < 0.4) {
        readRequestsCounter.add(1);
        const teamRes = http.get(`${BASE_URL}/users/team-invited`, { headers });

        const isRateLimit = isRateLimited(teamRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(teamRes, {
            'Team invited status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
        });

        if (!isRateLimit && Math.random() < 0.2) { // Reduced nested calls
            const level = Math.floor(Math.random() * 3) + 1;
            readRequestsCounter.add(1);
            const teamLevelRes = http.get(`${BASE_URL}/users/team-invited/${level}`, { headers });

            const levelRateLimit = isRateLimited(teamLevelRes);
            if (levelRateLimit) rateLimitCounter.add(1);

            check(teamLevelRes, {
                'Team level status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
            });
        }

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 1 + 0.8);

    // 6. Get Spin Prize List & Spin (30% of users - reduced from 40%)
    if (Math.random() < 0.3) {
        readRequestsCounter.add(1);
        const spinPrizeRes = http.get(`${BASE_URL}/spin-prize-list`, { headers });

        const isRateLimit = isRateLimited(spinPrizeRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(spinPrizeRes, {
            'Spin prize list status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
        });

        // Spin is a WRITE operation (Rate limit: 60/min = 1/second)
        if (!isRateLimit && Math.random() < 0.3) { // Reduced spin probability
            sleep(1); // Force spacing for write operations
            writeRequestsCounter.add(1);
            const spinRes = http.post(`${BASE_URL}/users/spin`, '{}', { headers });

            const spinRateLimit = isRateLimited(spinRes);
            if (spinRateLimit) rateLimitCounter.add(1);

            check(spinRes, {
                'Spin request completed or rate limited': (r) => [200, 400, 403, 429].includes(r.status) || isRateLimited(r),
            });

            if (spinRateLimit) sleep(Math.random() * 4 + 2);
        }

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 1 + 0.7);

    // Continue with other endpoints following similar pattern...
    // 7. Get Tasks (40% of users)
    if (Math.random() < 0.4) {
        readRequestsCounter.add(1);
        const tasksRes = http.get(`${BASE_URL}/users/task`, { headers });

        const isRateLimit = isRateLimited(tasksRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(tasksRes, {
            'Tasks status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
        });

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 1 + 0.6);

    // 8. Get Transactions with pagination (50% of users - reduced from 70%)
    if (Math.random() < 0.5) {
        const limit = [5, 10, 20][Math.floor(Math.random() * 3)]; // Reduced options
        const page = Math.floor(Math.random() * 3) + 1; // Reduced pages

        readRequestsCounter.add(1);
        const transactionsRes = http.get(`${BASE_URL}/users/transaction?limit=${limit}&page=${page}`, { headers });

        const isRateLimit = isRateLimited(transactionsRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(transactionsRes, {
            'Transactions status 200 or rate limited': (r) => r.status === 200 || isRateLimited(r),
            'Transactions response time reasonable': (r) => r.timings.duration < 1000,
        });

        if (isRateLimit) sleep(Math.random() * 3 + 1);
    }

    sleep(Math.random() * 2 + 1);

    // WRITE OPERATIONS (Rate limit: 60/min = 1/second)
    // 9. Update Bank Info (10% of users - reduced from 20%)
    if (Math.random() < 0.1) {
        const bankPayload = {
            id: 1,
            bank_id: Math.floor(Math.random() * 10) + 1,
            account_number: generateAccountNumber().toString(),
            account_name: generateAccountName()
        };

        writeRequestsCounter.add(1);
        const bankRes = http.put(`${BASE_URL}/users/bank`, JSON.stringify(bankPayload), { headers });

        const isRateLimit = isRateLimited(bankRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(bankRes, {
            'Bank update completed or rate limited': (r) => [200, 400, 422, 429].includes(r.status) || isRateLimited(r),
        });

        if (isRateLimit) sleep(Math.random() * 4 + 2);
    }

    sleep(Math.random() * 2 + 1);

    // 10. Create Investment (8% of users - reduced from 15%)
    if (Math.random() < 0.08) {
        const investmentPayload = {
            product_id: Math.floor(Math.random() * 5) + 1,
            amount: [1000, 5000, 10000][Math.floor(Math.random() * 3)],
            payment_method: 'QRIS'
        };

        writeRequestsCounter.add(1);
        const investmentRes = http.post(`${BASE_URL}/users/investments`, JSON.stringify(investmentPayload), { headers });

        const isRateLimit = isRateLimited(investmentRes);
        if (isRateLimit) rateLimitCounter.add(1);

        const investmentSuccess = check(investmentRes, {
            'Investment completed or rate limited': (r) => [200, 201, 400, 422, 429].includes(r.status) || isRateLimited(r),
        });

        // Check payment if investment created (READ operation)
        if (investmentSuccess && [200, 201].includes(investmentRes.status)) {
            sleep(2); // Spacing between write and read
            const invoiceId = `INV-${Math.floor(Math.random() * 9000000000) + 1000000000}`;

            readRequestsCounter.add(1);
            const paymentRes = http.get(`${BASE_URL}/users/payments/${invoiceId}`, { headers });

            const paymentRateLimit = isRateLimited(paymentRes);
            if (paymentRateLimit) rateLimitCounter.add(1);

            check(paymentRes, {
                'Payment check completed or rate limited': (r) => [200, 404, 429].includes(r.status) || isRateLimited(r),
            });
        }

        if (isRateLimit) sleep(Math.random() * 4 + 2);
    }

    sleep(Math.random() * 2 + 1);

    // 11. Change Password (3% of users - reduced from 5%)
    if (Math.random() < 0.03) {
        const changePasswordPayload = {
            current_password: '123456',
            password: '123456',
            confirmation_password: '123456'
        };

        writeRequestsCounter.add(1);
        const changePasswordRes = http.post(`${BASE_URL}/users/change-password`, JSON.stringify(changePasswordPayload), { headers });

        const isRateLimit = isRateLimited(changePasswordRes);
        if (isRateLimit) rateLimitCounter.add(1);

        check(changePasswordRes, {
            'Change password completed or rate limited': (r) => [200, 400, 422, 429].includes(r.status) || isRateLimited(r),
        });

        if (isRateLimit) sleep(Math.random() * 4 + 2);
    }

    // Longer sleep to respect overall rate limits
    sleep(Math.random() * 3 + 2);
}

export function handleSummary(data) {
    const totalRequests = data.metrics.http_reqs.values.count;
    const rateLimitHits = data.metrics.rate_limit_hits?.values.count || 0;
    const authRequests = data.metrics.auth_requests?.values.count || 0;
    const readRequests = data.metrics.read_requests?.values.count || 0;
    const writeRequests = data.metrics.write_requests?.values.count || 0;

    console.log('\n🔥 === K6 RATE LIMIT ANALYSIS SUMMARY ===');
    console.log(`✅ Total Requests: ${totalRequests}`);
    console.log(`📊 Avg Response Time: ${Math.round(data.metrics.http_req_duration.values.avg)}ms`);
    console.log(`🎯 95th Percentile: ${Math.round(data.metrics.http_req_duration.values['p(95)'])}ms`);
    console.log(`❌ Error Rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%`);
    console.log(`🚫 Rate Limit Hits: ${rateLimitHits} (${((rateLimitHits/totalRequests)*100).toFixed(1)}%)`);
    console.log('');
    console.log('📈 REQUEST BREAKDOWN:');
    console.log(`🔐 Auth Requests: ${authRequests} (Limit: 10/min)`);
    console.log(`👁️  Read Requests: ${readRequests} (Limit: 120/min)`);
    console.log(`✏️  Write Requests: ${writeRequests} (Limit: 60/min)`);
    console.log('');
    console.log('💡 RATE LIMIT ANALYSIS:');
    if (rateLimitHits > 0) {
        console.log(`⚠️  Rate limiting is ACTIVE and working as expected`);
        console.log(`📉 Consider optimizing client-side request spacing`);
        console.log(`🔄 Implement exponential backoff in production`);
    } else {
        console.log(`✅ No rate limits hit - API handled load well`);
    }

    console.log(`🔐 Avg Login Time: ${Math.round(data.metrics.login_duration?.values.avg || 0)}ms`);
    console.log(`📝 Avg Register Time: ${Math.round(data.metrics.register_duration?.values.avg || 0)}ms`);

    return {
        'rate-limit-analysis.json': JSON.stringify({
            summary: data,
            rate_limit_analysis: {
                total_requests: totalRequests,
                rate_limit_hits: rateLimitHits,
                rate_limit_percentage: ((rateLimitHits/totalRequests)*100).toFixed(1),
                auth_requests: authRequests,
                read_requests: readRequests,
                write_requests: writeRequests,
                avg_response_time: Math.round(data.metrics.http_req_duration.values.avg),
                p95_response_time: Math.round(data.metrics.http_req_duration.values['p(95)'])
            }
        }, null, 2),
    };
}
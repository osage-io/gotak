// GoTAK Load Testing with k6
// Comprehensive load testing scenarios for production readiness

import http from 'k6/http';
import ws from 'k6/ws';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const authFailureRate = new Rate('auth_failures');
const apiResponseTime = new Trend('api_response_time');
const wsConnectionTime = new Trend('ws_connection_time');
const totalRequests = new Counter('total_requests');

// Test configuration
const BASE_URL = __ENV.GOTAK_BASE_URL || 'http://localhost:8080';
const WS_URL = __ENV.GOTAK_WS_URL || 'ws://localhost:8080';

// Load test scenarios
export const options = {
  scenarios: {
    // Baseline load test - normal operations
    baseline_load: {
      executor: 'constant-arrival-rate',
      rate: 10, // 10 requests per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 5,
      maxVUs: 20,
      tags: { scenario: 'baseline' },
      exec: 'baselineTest',
    },
    
    // Stress test - high load
    stress_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 10,
      maxVUs: 100,
      stages: [
        { duration: '2m', target: 50 }, // Ramp up to 50 RPS
        { duration: '5m', target: 100 }, // Stay at 100 RPS
        { duration: '2m', target: 0 }, // Ramp down
      ],
      tags: { scenario: 'stress' },
      exec: 'stressTest',
    },
    
    // Spike test - sudden load increases
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 20,
      maxVUs: 200,
      stages: [
        { duration: '30s', target: 10 }, // Normal load
        { duration: '1m', target: 200 }, // Spike to 200 RPS
        { duration: '30s', target: 10 }, // Back to normal
      ],
      tags: { scenario: 'spike' },
      exec: 'spikeTest',
    },
    
    // WebSocket real-time test
    websocket_test: {
      executor: 'constant-vus',
      vus: 50,
      duration: '3m',
      tags: { scenario: 'websocket' },
      exec: 'websocketTest',
    },
    
    // Database intensive test
    database_test: {
      executor: 'constant-arrival-rate',
      rate: 20,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 10,
      maxVUs: 50,
      tags: { scenario: 'database' },
      exec: 'databaseTest',
    }
  },
  
  // Performance thresholds
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests under 1s
    http_req_failed: ['rate<0.01'], // Error rate under 1%
    api_response_time: ['p(90)<500'], // 90% of API responses under 500ms
    auth_failures: ['rate<0.05'], // Authentication failure rate under 5%
    ws_connection_time: ['p(95)<2000'], // WebSocket connections under 2s
  },
};

// Test data
const testUsers = [
  { username: 'test_user1', password: 'test123' },
  { username: 'test_user2', password: 'test123' },
  { username: 'test_admin', password: 'admin123' },
];

// Authentication helper
function authenticate(user) {
  const loginResponse = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    username: user.username,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const authSuccess = check(loginResponse, {
    'authentication successful': (r) => r.status === 200,
    'received auth token': (r) => JSON.parse(r.body).token !== undefined,
  });
  
  authFailureRate.add(!authSuccess);
  
  if (authSuccess) {
    return JSON.parse(loginResponse.body).token;
  }
  return null;
}

// Baseline load test - typical user operations
export function baselineTest() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const token = authenticate(user);
  
  if (!token) return;
  
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  group('User Dashboard Operations', function() {
    // Get user profile
    let response = http.get(`${BASE_URL}/api/user/profile`, { headers });
    check(response, {
      'profile loaded': (r) => r.status === 200,
      'profile response time OK': (r) => r.timings.duration < 500,
    });
    apiResponseTime.add(response.timings.duration);
    totalRequests.add(1);
    
    sleep(1);
    
    // List routes
    response = http.get(`${BASE_URL}/api/routes`, { headers });
    check(response, {
      'routes listed': (r) => r.status === 200,
      'routes response time OK': (r) => r.timings.duration < 1000,
    });
    apiResponseTime.add(response.timings.duration);
    totalRequests.add(1);
    
    sleep(2);
    
    // List geofences
    response = http.get(`${BASE_URL}/api/geofences`, { headers });
    check(response, {
      'geofences listed': (r) => r.status === 200,
      'geofences response time OK': (r) => r.timings.duration < 1000,
    });
    apiResponseTime.add(response.timings.duration);
    totalRequests.add(1);
  });
  
  sleep(3);
}

// Stress test - high load operations
export function stressTest() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const token = authenticate(user);
  
  if (!token) return;
  
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  group('High Load Operations', function() {
    // Multiple concurrent API calls
    const responses = http.batch([
      ['GET', `${BASE_URL}/api/routes`, null, { headers }],
      ['GET', `${BASE_URL}/api/geofences`, null, { headers }],
      ['GET', `${BASE_URL}/api/offline-areas`, null, { headers }],
      ['GET', `${BASE_URL}/api/tactical-overlays`, null, { headers }],
    ]);
    
    responses.forEach((response, index) => {
      check(response, {
        [`batch request ${index} successful`]: (r) => r.status === 200,
        [`batch request ${index} fast enough`]: (r) => r.timings.duration < 2000,
      });
      apiResponseTime.add(response.timings.duration);
      totalRequests.add(1);
    });
  });
  
  sleep(1);
}

// Spike test - sudden load handling
export function spikeTest() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const token = authenticate(user);
  
  if (!token) return;
  
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  group('Spike Load Test', function() {
    // Health check (most basic endpoint)
    const response = http.get(`${BASE_URL}/health`);
    check(response, {
      'health check responds': (r) => r.status === 200,
      'health check fast': (r) => r.timings.duration < 200,
    });
    apiResponseTime.add(response.timings.duration);
    totalRequests.add(1);
    
    // Quick API call
    const apiResponse = http.get(`${BASE_URL}/api/user/profile`, { headers });
    check(apiResponse, {
      'API responds under spike': (r) => r.status === 200,
      'API response time acceptable under spike': (r) => r.timings.duration < 3000,
    });
    apiResponseTime.add(apiResponse.timings.duration);
    totalRequests.add(1);
  });
}

// WebSocket test - real-time functionality
export function websocketTest() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const token = authenticate(user);
  
  if (!token) return;
  
  const wsUrl = `${WS_URL}/ws`;
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const startTime = Date.now();
  
  const response = ws.connect(wsUrl, params, function(socket) {
    const connectionTime = Date.now() - startTime;
    wsConnectionTime.add(connectionTime);
    
    socket.on('open', function() {
      console.log(`WebSocket connected for ${user.username}`);
      
      // Send position updates
      for (let i = 0; i < 10; i++) {
        socket.send(JSON.stringify({
          type: 'position_update',
          data: {
            lat: 39.0458 + (Math.random() - 0.5) * 0.01,
            lng: -76.6413 + (Math.random() - 0.5) * 0.01,
            timestamp: Date.now(),
            callsign: user.username,
          },
        }));
        sleep(0.5);
      }
      
      // Send chat messages
      socket.send(JSON.stringify({
        type: 'chat_message',
        data: {
          message: `Load test message from ${user.username}`,
          sender: user.username,
          timestamp: Date.now(),
        },
      }));
    });
    
    socket.on('message', function(message) {
      console.log(`WebSocket message received: ${message}`);
    });
    
    socket.on('error', function(error) {
      console.log(`WebSocket error: ${error}`);
    });
    
    // Keep connection alive for test duration
    sleep(10);
  });
  
  check(response, {
    'WebSocket connected successfully': (r) => r && r.url === wsUrl,
  });
}

// Database intensive test - CRUD operations
export function databaseTest() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  const token = authenticate(user);
  
  if (!token) return;
  
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
  
  group('Database Operations', function() {
    // Create a route
    const newRoute = {
      name: `Load Test Route ${Date.now()}`,
      description: 'Created during load testing',
      waypoints: [
        { lat: 39.0458, lng: -76.6413 },
        { lat: 39.0468, lng: -76.6423 },
      ],
      vehicle: 'car',
      routeType: 'fastest',
    };
    
    const createResponse = http.post(`${BASE_URL}/api/routes`, JSON.stringify(newRoute), { headers });
    const routeCreated = check(createResponse, {
      'route created': (r) => r.status === 201,
      'route creation time OK': (r) => r.timings.duration < 2000,
    });
    apiResponseTime.add(createResponse.timings.duration);
    totalRequests.add(1);
    
    if (routeCreated) {
      const routeId = JSON.parse(createResponse.body).id;
      sleep(1);
      
      // Read the route
      const readResponse = http.get(`${BASE_URL}/api/routes/${routeId}`, { headers });
      check(readResponse, {
        'route read': (r) => r.status === 200,
        'route read time OK': (r) => r.timings.duration < 1000,
      });
      apiResponseTime.add(readResponse.timings.duration);
      totalRequests.add(1);
      
      sleep(1);
      
      // Update the route
      const updatedRoute = { ...newRoute, description: 'Updated during load testing' };
      const updateResponse = http.put(`${BASE_URL}/api/routes/${routeId}`, JSON.stringify(updatedRoute), { headers });
      check(updateResponse, {
        'route updated': (r) => r.status === 200,
        'route update time OK': (r) => r.timings.duration < 2000,
      });
      apiResponseTime.add(updateResponse.timings.duration);
      totalRequests.add(1);
      
      sleep(1);
      
      // Delete the route
      const deleteResponse = http.del(`${BASE_URL}/api/routes/${routeId}`, null, { headers });
      check(deleteResponse, {
        'route deleted': (r) => r.status === 204,
        'route deletion time OK': (r) => r.timings.duration < 1000,
      });
      apiResponseTime.add(deleteResponse.timings.duration);
      totalRequests.add(1);
    }
  });
  
  sleep(2);
}

// Test teardown
export function teardown(data) {
  console.log('\n=== Load Test Summary ===');
  console.log(`Total requests made: ${totalRequests.count || 0}`);
  console.log('Test completed successfully');
}

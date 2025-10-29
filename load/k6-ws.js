import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const wsConnectionSuccess = new Rate('ws_connection_success');
const messageDeliverySuccess = new Rate('message_delivery_success');

export let options = {
  stages: [
    { duration: '2m', target: 1000 },   // Ramp up to 1k users
    { duration: '3m', target: 5000 },   // Ramp up to 5k users
    { duration: '5m', target: 10000 },  // Ramp up to 10k users
    { duration: '5m', target: 10000 },  // Stay at 10k users
    { duration: '2m', target: 0 },      // Ramp down
  ],
  thresholds: {
    'ws_connection_success': ['rate>0.95'],
    'message_delivery_success': ['rate>0.95'],
    'http_req_duration': ['p(99)<200'], // 99% of requests should be below 200ms
  },
};

// Get JWT from environment or generate one
const JWT = __ENV.JWT || generateTestJWT();

function generateTestJWT() {
  // In a real test, you'd call the /auth/register or /auth/login endpoint
  // For now, return a placeholder
  console.warn('No JWT provided. Set JWT environment variable.');
  return 'test-jwt';
}

export default function () {
  const url = `ws://localhost:8080/v1/ws?token=${JWT}&device=k6-${__VU}`;

  const response = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      console.log(`VU ${__VU}: Connected`);
      wsConnectionSuccess.add(1);

      // Send a message every 5 seconds
      socket.setInterval(() => {
        const message = JSON.stringify({
          type: 'SendMessage',
          uuid: `${__VU}-${__ITER}-${Date.now()}`,
          chatId: 1,
          body: `Load test message from VU ${__VU} iteration ${__ITER}`,
        });
        socket.send(message);
      }, 5000);

      // Send ping every 30 seconds
      socket.setInterval(() => {
        socket.send(JSON.stringify({
          type: 'Ping',
          timestamp: Date.now(),
        }));
      }, 30000);
    });

    socket.on('message', (data) => {
      const msg = JSON.parse(data);

      if (msg.type === 'Delivered') {
        messageDeliverySuccess.add(1);
        check(msg, {
          'delivered has uuid': (m) => m.uuid !== undefined,
          'delivered has msgId': (m) => m.msgId !== undefined,
        });
      } else if (msg.type === 'Message') {
        check(msg, {
          'message has body': (m) => m.body !== undefined,
          'message has chatId': (m) => m.chatId !== undefined,
        });
      } else if (msg.type === 'Pong') {
        // Pong received
      }
    });

    socket.on('close', () => {
      console.log(`VU ${__VU}: Disconnected`);
      wsConnectionSuccess.add(0);
    });

    socket.on('error', (e) => {
      console.error(`VU ${__VU}: Error: ${e.error()}`);
      wsConnectionSuccess.add(0);
    });

    // Keep connection alive for 60 seconds
    socket.setTimeout(() => {
      socket.close();
    }, 60000);
  });

  check(response, {
    'status is 101': (r) => r && r.status === 101,
  });

  // Small sleep between iterations
  sleep(1);
}

export function setup() {
  console.log('Starting WebSocket load test...');
  console.log('Make sure the gateway is running at localhost:8080');
  console.log('JWT token:', JWT ? 'Provided' : 'Missing');
}

export function teardown(data) {
  console.log('Load test completed');
}

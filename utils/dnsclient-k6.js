import http from 'k6/http';

export let options = {
  // Set the target URL of your API (Version A)
  url: 'http://go-grpc-defaultclient-with-envoy.demo/try',

  // Set the number of virtual users (VUs)
  vus: 10,

  // Set the duration of the test
  duration: '30s',

  // Optional: Define additional parameters
//   thresholds: {
//     http_req_failed: ['rate<0.01'], // Allow less than 1% failed requests
//     http_req_duration: ['p(95)<200'], // 95% of requests should take less than 200ms
//   },
};

export default function() {
  // Send a request to your API (Version A)
  // Adapt this based on your API endpoints and functionalities
  http.get(options.url);
}
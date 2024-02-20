import http from 'k6/http';

export let options = {
  // default
  vus: 10,
  // default
  duration: '30s',

};

export default function() {
  http.get('http://go-grpc-defaultclient-with-envoy.demo/try');
}
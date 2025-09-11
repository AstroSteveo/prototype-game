const test = require('node:test');
const assert = require('node:assert');
const request = require('supertest');
const app = require('../server');

test('GET /healthz returns 200', async () => {
  const res = await request(app).get('/healthz');
  assert.strictEqual(res.statusCode, 200);
});

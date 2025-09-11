const express = require('express');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 3000;
const DIST_DIR = path.join(__dirname, 'dist');

app.use(express.static(DIST_DIR));

app.get('/healthz', (_req, res) => {
  res.sendStatus(200);
});

app.use((_req, res) => {
  res.sendFile(path.join(DIST_DIR, 'index.html'));
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Server listening on port ${PORT}`);
  });
}

module.exports = app;

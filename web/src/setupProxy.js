const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  app.use('/pac-go-server', createProxyMiddleware({
    target: process.env.REACT_APP_PAC_GO_SERVER_TARGET || 'http://localhost:8000',
    changeOrigin: true,
    pathRewrite: {
      '^/pac-go-server': '/api/v1', //rewrite the path
    },
  }));
};

var path = require('path');

var express = require('express');
var app = express();


var index = path.join(__dirname, 'app.html')
app.get('/', function (req, res) {
  res.sendFile(index);
});

app.use(express.static('assets'));

var port = process.env.PORT || 3000;
var server = app.listen(port, function () {
  var host = server.address().address;

  console.log('Example app listening at http://%s:%s', host, port);
});

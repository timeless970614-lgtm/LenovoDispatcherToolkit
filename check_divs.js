var fs = require('fs');
var lines = fs.readFileSync('C:\\LenovoDispatcherToolkit\\frontend\\src\\pages\\FunctionCheck.vue', 'utf8').split('\n');
var depth = 0;
for (var i = 462; i < 775; i++) {
  var line = lines[i];
  var opens = (line.match(/<div[\s>]/g) || []).length;
  var closes = (line.match(/<\/div>/g) || []).length;
  depth += opens - closes;
  if (opens !== closes || depth <= 0) {
    console.log('L' + (i+1) + ' depth=' + depth + ' opens=' + opens + ' closes=' + closes + ' | ' + line.trim().substring(0, 80));
  }
}

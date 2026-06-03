var fs = require('fs');
var c = fs.readFileSync('C:\\LenovoDispatcherToolkit\\frontend\\src\\pages\\FunctionCheck.vue', 'utf8');
c = c.replace(/activeTab === 'b'/g, "activeTab === 'power'");
fs.writeFileSync('C:\\LenovoDispatcherToolkit\\frontend\\src\\pages\\FunctionCheck.vue', c);
console.log('done');

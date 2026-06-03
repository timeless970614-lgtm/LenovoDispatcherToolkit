var fs = require('fs');
var c = fs.readFileSync('C:\\LenovoDispatcherToolkit\\frontend\\src\\pages\\FunctionCheck.vue', 'utf8');
var lines = c.split('\n');

// Find the first ppm-container (line with "PPM Power Settings")
var startIdx = -1;
var endIdx = -1;
for (var i = 0; i < lines.length; i++) {
  if (startIdx === -1 && lines[i].includes('PPM Power Settings')) {
    // Go back to find the ppm-container opening div
    for (var j = i; j >= 0; j--) {
      if (lines[j].includes('ppm-container')) { startIdx = j; break; }
    }
  }
  if (startIdx !== -1 && endIdx === -1 && lines[i].includes('Real-time Power Consumption Section')) {
    // Go back to find the closing </div> before this comment
    for (var j = i - 1; j >= startIdx; j--) {
      if (lines[j].trim() === '</div>') { endIdx = j; break; }
    }
    break;
  }
}

console.log('Removing lines', startIdx + 1, 'to', endIdx + 1);
console.log('First line:', lines[startIdx].trim());
console.log('Last line:', lines[endIdx].trim());

// Remove lines startIdx..endIdx
lines.splice(startIdx, endIdx - startIdx + 1);

fs.writeFileSync('C:\\LenovoDispatcherToolkit\\frontend\\src\\pages\\FunctionCheck.vue', lines.join('\n'));
console.log('done');

(function(){"use strict";self.onmessage=e=>{if(e.data.type==="recognize"){const{width:t,height:s}=e.data.imageData,i={type:"result",blocks:[{id:"ocr-1",text:"极致体验",bbox:{x:t*.3,y:s*.4,width:t*.4,height:s*.1},confidence:.92}]};self.postMessage(i)}}})();
//# sourceMappingURL=ocr.worker-DQdUNs94.js.map

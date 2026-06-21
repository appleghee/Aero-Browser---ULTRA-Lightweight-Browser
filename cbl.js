(function(){
if(window.__mbCBL)return;
var C=window.__mbCBL={_def:0,_pri:0};
var _fi=window.fetch;
window.fetch=function(u,o){
var url=typeof u==='string'?u:u&&u.url?u.url:'';
var low=/\.(png|jpg|webp|gif|svg|avif|mp4|woff2?)/i.test(url);
var high=/\.(css|js|html?)$/i.test(url)||url.includes('api/');
if(low&&!high){
C._def++;
return new Promise(function(r){
setTimeout(function(){r(_fi.call(this,u,o));},100);
});
}
if(high)C._pri++;
return _fi.call(this,u,o);
};
})();

(function(){if(window.__mbPrf)return;
var w=window,d=document;
w.__mbPrf={};
var b=w.__mbBase||'http://127.0.0.1:'+(w.__mbPort||6969),t=w.__mbToken||'';
var h={'Content-Type':'application/json','X-API-Token':t};
var dn=w.location.hostname,ph={},phTimer=null;

function fullHref(el){
 if(el.href)return el.href;
 var hr=el.getAttribute('href');
 if(!hr)return '';
 try{return new URL(hr,w.location.href).href}catch(e){return ''}
}

function api(url,body,cb){
 var x=new XMLHttpRequest();
 x.open('POST',b+url,true);
 for(var k in h)x.setRequestHeader(k,h[k]);
 if(cb)x.onload=function(){try{cb(JSON.parse(x.responseText))}catch(e){}};
 x.send(JSON.stringify(body));
}

function doPrefetch(url){
 if(ph[url])return;
 ph[url]=1;
 var link=d.createElement('link');
 link.rel='prefetch';link.href=url;
 d.head.appendChild(link);
 var x=new XMLHttpRequest();
 x.open('GET',url,true);
 x.send();
}

d.addEventListener('mouseover',function(e){
 var el=e.target.closest('a');
 if(!el)return;
 var hr=fullHref(el);
 if(!hr||hr.indexOf('http')!==0)return;
 if(phTimer)clearTimeout(phTimer);
 phTimer=setTimeout(function(){
  phTimer=null;
  api('/api/predict/hover',{domain:dn,href:hr},function(r){
   if(r&&r.prefetch)doPrefetch(hr);
  });
 },200);
},true);

d.addEventListener('mouseout',function(e){
 var el=e.target.closest('a');
 if(!el||!phTimer)return;
 clearTimeout(phTimer);phTimer=null;
},true);

d.addEventListener('click',function(e){
 var el=e.target.closest('a');
 if(!el)return;
 var hr=fullHref(el);
 if(!hr||hr.indexOf('http')!==0)return;
 if(ph[hr]){api('/api/predict/hit',{domain:dn,href:hr})}
 else{api('/api/predict/click',{domain:dn,href:hr})}
},true);
})();

"use strict";(self.webpackChunkmemphis_ui=self.webpackChunkmemphis_ui||[]).push([[3047],{75525:function(e){function n(e){!function(e){function n(e,n){return"___"+e.toUpperCase()+n+"___"}Object.defineProperties(e.languages["markup-templating"]={},{buildPlaceholders:{value:function(t,a,r,o){if(t.language===a){var i=t.tokenStack=[];t.code=t.code.replace(r,(function(e){if("function"===typeof o&&!o(e))return e;for(var r,u=i.length;-1!==t.code.indexOf(r=n(a,u));)++u;return i[u]=e,r})),t.grammar=e.languages.markup}}},tokenizePlaceholders:{value:function(t,a){if(t.language===a&&t.tokenStack){t.grammar=e.languages[a];var r=0,o=Object.keys(t.tokenStack);!function i(u){for(var c=0;c<u.length&&!(r>=o.length);c++){var s=u[c];if("string"===typeof s||s.content&&"string"===typeof s.content){var p=o[r],l=t.tokenStack[p],g="string"===typeof s?s:s.content,f=n(a,p),k=g.indexOf(f);if(k>-1){++r;var m=g.substring(0,k),h=new e.Token(a,e.tokenize(l,t.grammar),"language-"+a,l),d=g.substring(k+f.length),y=[];m&&y.push.apply(y,i([m])),y.push(h),d&&y.push.apply(y,i([d])),"string"===typeof s?u.splice.apply(u,[c,1].concat(y)):s.content=y}}else s.content&&i(s.content)}return u}(t.tokens)}}}})}(e)}e.exports=n,n.displayName="markupTemplating",n.aliases=[]}}]);
import fs from 'fs';
const s=fs.readFileSync('app.js','utf8');
const a=s.indexOf('oo={skip:'), b=s.indexOf(',Bu={en:oo');
const src='const '+s.slice(a,b)+';export default {en:oo,tr:by,de:Xy,ru:Zy};';
fs.writeFileSync('dicts.mjs',src);
const d=(await import('./dicts.mjs')).default;
const en=Object.keys(d.en);
const rep={total_en_keys:en.length};
for(const l of ['tr','de','ru']){
  const own=Object.keys(d[l]);  // spread includes en keys; need own literal keys
  rep[l]={};
}
fs.writeFileSync('dicts.json',JSON.stringify(d,null,1));
console.log(rep);

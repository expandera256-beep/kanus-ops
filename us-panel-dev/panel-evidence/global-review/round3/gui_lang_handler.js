function loadLocale(lang) {
  const q = lang ? "?lang=" + encodeURIComponent(lang) : "";
  return fetch("/api/i18n" + q, { cache: "no-store" }).then(r => r.ok ? r.json() : null).then(doc => { if (doc) applyLocale(doc); }).catch(() => {});
}
// Persist the choice first; only switch when the server saved it. On any
// failure keep the previous language, restore the selector and say so.
function saveLanguage(lang) {
  return fetch("/api/lang", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ lang }) })
    .then(r => r.ok).catch(() => false);
}
function showLangSaveError(lang) {
  const el = document.querySelector("#langstatus");
  return fetch("/api/i18n?lang=" + encodeURIComponent(lang), { cache: "no-store" })
    .then(r => r.ok ? r.json() : null).catch(() => null)
    .then(doc => {
      const msg = doc && doc.messages && doc.messages["gui.lang.save_failed"];
      el.textContent = typeof msg === "string" ? msg : "Dil tercihi kaydedilemedi. / Could not save the language preference.";
      el.hidden = false;
    });
}
document.querySelector("#langsel").addEventListener("change", ev => {
  const sel = ev.target;
  const prev = i18nState.lang;
  const lang = sel.value;
  document.querySelector("#langstatus").hidden = true;
  return saveLanguage(lang).then(saved => {
    if (saved) return loadLocale(lang);
    sel.value = prev;
    return showLangSaveError(prev);
  });
});
loadLocale("");
load();
</script>
</body>
</html>

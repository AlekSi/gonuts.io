function bindEvent(el, e, fn) {
  if (el.addEventListener){
    el.addEventListener(e, fn, false);
  } else if (el.attachEvent){
    el.attachEvent('on'+e, fn);
  }
}

function bindSearchEvents() {
  var search = document.getElementById('search');
  function clearInactive() {
    if (search.className == "inactive") {
      search.value = "";
      search.className = "";
    }
  }
  function restoreInactive() {
    if (search.value !== "") {
      return;
    }
    if (search.type != "search") {
      search.value = search.getAttribute("placeholder");
    }
    search.className = "inactive";
  }
  restoreInactive();
  bindEvent(search, 'focus', clearInactive);
  bindEvent(search, 'blur', restoreInactive);
}

bindEvent(window, 'load', bindSearchEvents);

function dance(id, min, max) {
  var e = document.getElementById(id);
  if (e == null) {
    return;
  }
  setInterval(function() {
    e.textContent = Math.floor(Math.random() * (max - min + 1)) + min;
  }, 2000)
}

function startFun() {
  dance("version-count", 50000, 100000)
  dance("nuts-count", 5000, 10000)
  dance("users-count", 10000, 50000)
}

bindEvent(window, 'load', startFun);

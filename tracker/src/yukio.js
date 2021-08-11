(function () {
  'use strict';

  var location = window.location
  var document = window.document


  var scriptEl = document.currentScript;
  var endpoint = scriptEl.getAttribute('data-api') || defaultEndpoint(scriptEl)
  var yukio_ignore = window.localStorage.yukio_ignore;
  <<if .exclusions >>
  var excludedPaths = scriptEl && scriptEl.getAttribute('data-exclude').split(',');
  << end >>
  var lastPage;

  function warn(reason) {
    console.warn('Ignoring Event: ' + reason);
  }

  function defaultEndpoint(el) {
    return new URL(el.src).origin + '/api/event'
  }


  function trigger(eventName, options) {
    if (/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(location.hostname) || location.protocol === 'file:') return warn('localhost');
    if (window.phantom || window._phantom || window.__nightmare || window.navigator.webdriver || window.Cypress) return;
    if (yukio_ignore == "true") return warn('localStorage flag')
      <<if .exclusions >>
    if (excludedPaths)
        for (var i = 0; i < excludedPaths.length; i++)
          if (eventName == "pageview" && location.pathname.match(new RegExp('^' + excludedPaths[i].trim().replace(/\*\*/g, '.*').replace(/([^\.])\*/g, '$1[^\\s\/]*') + '\/?$')))
            return warn('exclusion rule');
    << end >>

    var payload = {}
    payload.n = eventName
    payload.u = location.href
    payload.d = scriptEl.getAttribute('data-domain')
    payload.r = document.referrer || null
    payload.w = window.innerWidth
    if (options && options.meta) {
      payload.m = JSON.stringify(options.meta)
    }
    if (options && options.props) {
      payload.p = JSON.stringify(options.props)
    }
    <<if .hash >>
      payload.h = 1
        << end >>

    var request = new XMLHttpRequest();
    request.open('POST', endpoint, true);
    request.setRequestHeader('Content-Type', 'text/plain');

    request.send(JSON.stringify(payload));

    request.onreadystatechange = function () {
      if (request.readyState == 4) {
        options && options.callback && options.callback()
      }
    }
  }

  function page() {
    <<if  .hash | not >>
    if (lastPage === location.pathname) return;
    << end >>
      lastPage = location.pathname
    trigger('pageview')
  }

  function handleVisibilityChange() {
    if (!lastPage && document.visibilityState === 'visible') {
      page()
    }
  }

  <<if .outbound_links >>
    function handleOutbound(event) {
      var link = event.target;
      var middle = event.type == "auxclick" && event.which == 2;
      var click = event.type == "click";
      while (link && (typeof link.tagName == 'undefined' || link.tagName.toLowerCase() != 'a' || !link.href)) {
        link = link.parentNode
      }

      if (link && link.href && link.host && link.host !== location.host) {
        if (middle || click)
          yukio('Outbound Link: Click', { props: { url: link.href } })

        // Delay navigation so that yukio is notified of the click
        if (!link.target || link.target.match(/^_(self|parent|top)$/i)) {
          if (!(event.ctrlKey || event.metaKey || event.shiftKey) && click) {
            setTimeout(function () {
              location.href = link.href;
            }, 150);
            event.preventDefault();
          }
        }
      }
    }

  function registerOutboundLinkEvents() {
      document.addEventListener('click', handleOutbound)
      document.addEventListener('auxclick', handleOutbound)
    }
  << end >>

    <<if .hash >>
    window.addEventListener('hashchange', page)
    <<else>>
  var his = window.history
  if (his.pushState) {
    var originalPushState = his['pushState']
    his.pushState = function () {
      originalPushState.apply(this, arguments)
      page();
    }
    window.addEventListener('popstate', page)
  }
  << end >>

    <<if .outbound_links >>
    registerOutboundLinkEvents()
    << end >>

  var queue = (window.yukio && window.yukio.q) || []
  window.yukio = trigger
  for (var i = 0; i < queue.length; i++) {
    trigger.apply(this, queue[i])
  }

  if (document.visibilityState === 'prerender') {
    document.addEventListener("visibilitychange", handleVisibilityChange);
  } else {
    page()
  }
})();

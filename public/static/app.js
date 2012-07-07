$(document).ready(function() {
  var host = $('meta[name=host]').attr('content');
  $('#email').change(function() {
    var script = "javascript:(function() { \
      var setupDiv = function() { \
        var id = 'Tinderizer'; \
        var div = document.getElementById(id); \
        var body = document.getElementsByTagName('body')[0]; \
        if (null != div) { \
          body.removeChild(div); \
        } \
        div = document.createElement('div'); \
        div.id = id; div.style.width = 'auto'; div.style.height = '30px'; \
        div.style.fontSize = '14px'; div.setAttribute('data-host', '" + host + "'); \
        div.setAttribute('data-email', '" + encodeURIComponent(this.value) + "'); \
        div.style.position = 'fixed'; div.style.top = '10px'; div.style.left = '10px'; \
        div.style.background = 'white'; div.style.color = 'black'; div.style.borderColor = 'black'; div.style.borderStyle = 'solid'; \
        div.style.borderWidth = '2px'; div.style.zIndex = '99999999'; div.style.padding = '16px'; \
        div.style.textAlign = 'center'; \
        div.innerHTML = 'Working...'; \
        body.appendChild(div); \
      }; \
      setupDiv(); \
      var script = document.createElement('script'); \
      script.type = 'text/javascript'; \
      script.src = 'http://" + host + "/static/bookmarklet.js?t=' + (new Date()).getTime(); \
      document.getElementsByTagName('head')[0].appendChild(script); \
    })();";
    $('#bookmarklet').attr('href', script);
  });

  $(document).bind('reveal.facebox', function() {
    $('#ios').html($('#bookmarklet').attr('href'));
  });

  $.facebox.settings.closeImage = '/static/closelabel.png';
  $.facebox.settings.loadingImage = '/static/loading.gif';

  $('.facebox').facebox();

  $('.slidedeck').slidedeck().vertical();
  $('a.vsnext').click(function() {
    $('.slidedeck').slidedeck().vertical().next();
    return false;
  });
  $('a.hsnext').click(function() {
    $('.slidedeck').slidedeck().next();
    return false;
  });
});

((url) ->
  request = if 'XDomainRequest' of window
    (url, success) ->
      xdr = new XDomainRequest()
      xdr.onload = () ->
        success(JSON.parse(xdr.responseText))
      xdr.open('GET', url)
      xdr.send(null)
  else
    (url, success) ->
      xhr = new XMLHttpRequest()
      xhr.onreadystatechange = () ->
        if xhr.readyState == 4
          success(JSON.parse(xhr.responseText))
      xhr.open('GET', url, true)
      xhr.setRequestHeader('Accept', 'application/json')
      xhr.send(null)

  escapeRegex = (text) ->
    text.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, "\\$&")

  body = document.getElementsByTagName('body')[0]
  div = document.getElementById('Tinderizer') || document.getElementById('kindlebility')
  host = div.getAttribute('data-host')
  to = div.getAttribute('data-email')
  redirect = false
  notify = (message) ->
    div.innerHTML = message
    div.appendChild(document.createTextNode(' '))

  div.style.width = '300px'
  div.style.height = '30px'
  div.style.fontSize = '12px'

  # TODO: Some sort of detection of a failure
  Tinderizer = () ->
    validHost = /tinderizer\.com/i
    if !validHost.test(host)
      if confirm("Kindlebility has been renamed to Tinderizer. Please remake your bookmark to ensure it continues to work after the domain completely changes!\n\nPlease click OK to visit the new website and remake your bookmarklet when we're done here.")
        redirect = true

    params = "?url=#{encodeURIComponent(url)}&email=#{encodeURIComponent(to)}&t=#{(new Date()).getTime()}"
    request "http://#{host}/ajax/submit.json" + params, (submit) ->
      notify(submit.message)
      if submit.limited || !submit.id
        setTimeout((() ->
          body.removeChild(div)
        ), 2500)
        return

      id = submit.id
      timer = setInterval((() ->
        request "http://#{host}/ajax/status/#{id}.json?t=#{(new Date()).getTime()}", (status) ->
          notify(status.message)
          if status.done
            clearInterval(timer)
            setTimeout((() ->
              body.removeChild(div)
              window.location = 'http://tinderizer.com/' if redirect
            ), 2500)
      ), 500)

  checks = {
    # "You need to run this on an article page! Main or home pages don't work very well.": new RegExp(escapeRegex(window.location.protocol + "//#{window.location.host}/") + '$'),
    'There is nothing to do on about:blank!': /about:blank/,
    'You need to run this on a publicly accessible HTML page!': /\.(pdf|jpg)$/i,
    'Run this on the raw page, not a Readability page!': /^https?:\/\/www.readability.com\/articles\//i
  }

  for own message, regex of checks
    if regex.test(url)
      alert(message)
      body.removeChild(div)
      return

  Tinderizer()
)(document.location.href)

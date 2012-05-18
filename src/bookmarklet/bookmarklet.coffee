((url) ->
  interval = (time, func) ->
    setInterval(func, time)

  timeout = (time, func) ->
    setTimeout(func, time)

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

  div.style.minWidth = '300px'
  div.style.width = 'auto'
  div.style.height = '30px'
  div.style.fontSize = '12px'
  div.style.fontFamily = 'sans-serif'

  Tinderizer = () ->
    validHost = /tinderizer\.com/i
    unless validHost.test(host)
      if confirm("Kindlebility has been renamed to Tinderizer. Please remake your bookmark to ensure it continues to work after the domain completely changes!\n\nPlease click OK to visit the new website and remake your bookmarklet when we're done here.")
        redirect = true

    params = "?url=#{encodeURIComponent(url)}&email=#{encodeURIComponent(to)}&t=#{(new Date()).getTime()}"
    request "http://#{host}/ajax/submit.json#{params}", (submit) ->
      notify(submit.message)
      if submit.limited || !submit.id?
        timeout(2500, -> body.removeChild(div))
        return

      done = false
      broken = timeout 30000, ->
        # If we can't accomplish stuff in 30 seconds, something is borked.
        done = true
        alert("Okay, this is getting out of hand, something must have broken, I'm going to stop trying.")
        body.removeChild(div)

      id = submit.id
      timer = interval 500, ->
        clearInterval(timer) if done
        request "http://#{host}/ajax/status/#{id}.json?t=#{(new Date()).getTime()}", (status) ->
          notify(status.message)
          if status.done
            done = true
            clearTimeout(broken)
            clearInterval(timer)
            timeout 2500, ->
              body.removeChild(div)
              window.location = 'http://tinderizer.com/' if redirect

  checks = {
    "This doesn't work on local files.": /^file:/
    "You need to run this on an article page! Main or home pages don't work very well.": new RegExp(escapeRegex(window.location.protocol + "//#{window.location.host}/") + '$')
    'There is nothing to do on about:blank!': /about:blank/
    'You need to run this on a publicly accessible HTML page!': /\.(pdf|jpg)$/i
    'Run this on the raw page, not a Readability page!': /^https?:\/\/www.readability.com\/articles\//i
  }

  for own message, regex of checks
    if regex.test(url)
      alert(message)
      body.removeChild(div)
      return

  Tinderizer()
)(document.location.href)

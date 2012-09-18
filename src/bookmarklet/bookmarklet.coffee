interval = (time, func) -> setInterval(func, time)
timeout = (time, func) -> setTimeout(func, time)
escapeRegex = (text) -> text.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, "\\$&")

Request = if 'XDomainRequest' of window
    (url, method, data, success) ->
        xdr = new XDomainRequest()
        xdr.onload = -> success(JSON.parse(xdr.responseText))
        xdr.open(method, url)
        xdr.send(data)
else
    (url, method, data, success) ->
        xhr = new XMLHttpRequest()
        xhr.onreadystatechange = ->
            if xhr.readyState == 4
                success(JSON.parse(xhr.responseText))
        xhr.open(method, url, true)
        if data?
            xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded')
        xhr.setRequestHeader('Accept', 'application/json')
        xhr.send(data)

class Tinderizer
    paywalls: /arstechnica|nytimes|theatlantic|guardian|wsj|thetimes/

    css: {{.Style}}

    validHost: /tinderizer\.com/i

    checks:
        "This doesn't work on local files.": /^file:/
        "You need to run this on an article page! Main or home pages don't work very well.": new RegExp(escapeRegex(window.location.protocol + "//#{window.location.host}/") + '$')
        'There is nothing to do on about:blank!': /about:blank/
        'You need to run this on a publicly accessible HTML page!': /\.(pdf|jpg)$/i
        'Run this on the raw page, not a Readability page!': /^https?:\/\/www.readability.com\/articles\//i

    constructor: (@div, @url) ->
        @host = String(@div.getAttribute('data-host')).split(':')[0]
        @to = @div.getAttribute('data-email')
        @body = document.getElementsByTagName('body')[0]
        @redirect = false
        @submitEndpoint = "{{.Protocol}}://#{@host}/ajax/submit.json"

    okay: ->
        for own message, regex of @checks
            if regex.test(@url)
                alert(message)
                @body.removeChild(@div)
                return false
        return true

    notify: (message) ->
        @div.innerHTML = message
        @div.appendChild(document.createTextNode(' '))

    appendStyleSheet: ->
        head = document.getElementsByTagName('head')[0]
        if head?
            style = document.createElement('style')
            style.type = 'text/css'
            head.appendChild(style)
            if 'styleSheet' of style
                style.styleSheet.cssText = @css
            else
                style.appendChild(document.createTextNode(@css))

    checkHost: ->
        unless @validHost.test(@host)
            if confirm("Kindlebility has been renamed to Tinderizer. Please remake your bookmark to ensure it continues to work after the domain completely changes!\n\nPlease click OK to visit the new website and remake your bookmarklet when we're done here.")
                @redirect = true

    onSubmit: (data) =>
        @notify(data.message)
        if data.limited || !data.id?
            timeout(2500, -> @body.removeChild(@div))
            return

        @done = false
        broken = timeout 30000, =>
            # If we can't accomplish stuff in 30 seconds, something is borked.
            @done = true
            alert("Okay, this is getting out of hand, something must have broken, I'm going to stop trying.")
            @body.removeChild(@div)

        id = data.id
        timer = interval 500, =>
            clearInterval(timer) if @done
            Request "{{.Protocol}}://#{@host}/ajax/status/#{id}.json?t=#{(new Date()).getTime()}", 'GET', null, (status) =>
                @notify(status.message)
                if status.done
                    @done = true
                    clearTimeout(broken)
                    clearInterval(timer)
                    timeout 2500, =>
                        @body.removeChild(@div)
                        window.location = 'https://tinderizer.com/' if @redirect

    isPaywall: ->
        @paywalls.test(document.location.host)

    run: ->
        if @okay()
            @appendStyleSheet()
            @checkHost()
            data =
                url: @url
                email: @to
            if @isPaywall()
                data.content = document.documentElement.outerHTML
            Request(@submitEndpoint, 'POST', JSON.stringify(data), @onSubmit)

div = document.getElementById('Tinderizer') || document.getElementById('kindlebility')
url = document.location.href
tinderizer = new Tinderizer(div, url)
tinderizer.run()

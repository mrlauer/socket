$( ->
    write = (msg) ->
        $('#out').append $("<p>#{msg}</p>")

    write "Starting..."

#     url = "ws://echo.websocket.org/"
    url = "ws://#{document.location.host}/news"

    makeWS = (tag)->
        ws = new WebSocket url
        ws.onopen = ->
            write "Websocket Opened"
            sendJson ws, "This is our first message!"
        ws.onmessage = (evt)->
            write "#{tag} Got message #{evt.data}"
        ws.onerror = (data)->
            write "Got error #{data}"
        ws.onclose = ->
            write "Closed"
        return ws

    ws1 = makeWS 'WS1'
    ws2 = makeWS 'WS2'

    $('#out').click ->
        sendJson ws1, 'Click!'
        sendJson ws2, 'Also click'

    $('#close').click ->
        ws1.close()

    hex = (n) ->
        result = ""
        while n > 0
            d = n & 0xf
            n = (n >> 4)
            result = '0123456789abcdef'[d] + result
        return result

    sendJson = (ws, o) ->
        js = JSON.stringify { data : o }
        len = js.length
        slen = hex len
        while slen.length < 8
            slen = "0" + slen
        ws.send slen + js

    write "URL is #{document.location.host}"
)

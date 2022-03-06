function eID(id) { return document.getElementById(id) }

async function init() {
    const useripinfofetch = await fetch("http://" + window.location.host + "/ipinfo", {method: "post"})
    const useripinfo = useripinfofetch.ok ? await useripinfofetch.json() : {ok: false}
    const uip = eID("uip")
    
    // internal OK, not fetch OK
    if (!useripinfo.ok) {
        eID("info").innerHTML += "<p>Sorry, an internal error has occurred. Please try again later.</p>"
        uip.innerHTML = uip.innerHTML.replace("{{userip}}", "Error")
        return
    }
    uip.innerHTML = uip.innerHTML.replace("{{userip}}", useripinfo.ip)
}

init()
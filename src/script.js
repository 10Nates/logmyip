function eID(id) { return document.getElementById(id) }

async function init() {
    const useripinfofetch = await fetch(window.location.hostname + "/ipinfo", {method: "POST"})
    const useripinfo = useripinfofetch.ok ? await useripinfofetch.json() : {ok: false}
    if (!useripinfo.ok) {
        eID("info") += "<p>Sorry, an internal error has occurred. Please try again later.</p>"
    }
    const uip = eID("uip")
    uip.innerHTML = uip.innerHTML.replace("{{userip}}", useripinfo.ip)
}
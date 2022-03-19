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

function logUserIP() {
    // silent form handling
    fetch("/logip", {
        method: "POST",
        headers: {'Content-Type': 'application/x-www-form-urlencoded'}, 
        body: "confirm=yes"
      }).then(res => {
        switch (res.status) {
            case 200:
                eID("formreturn").innerHTML = "IP successfully logged!"
                break;
            default:
                eID("formreturn").innerHTML = "Error logging IP. Please try again later."
                break;

        }
      });
}

function silenceForm() {
    var form = document.getElementById("ipform");
    function handleForm(event) { 
        event.preventDefault(); 
        logUserIP()
    } 
    form.addEventListener('submit', handleForm);
}

silenceForm()
init()
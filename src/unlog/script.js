function eID(id) { return document.getElementById(id) }

if (isLogged==="yes") {
    eID("status").innerHTML = eID("status").innerHTML.replace("LOADING...", "logged.")
    eID("removeField").style.display = "inherit"
} else if (isLogged==="no") {
    eID("status").innerHTML = eID("status").innerHTML.replace("LOADING...", "NOT logged.")
} else {
    eID("status").innerHTML = eID("status").innerHTML.replace("LOADING...", "... an error has occured. Please try again later.")
}

function unlogForm() {
    // silent href handling (it's an href this time because it's funnier)
    fetch("/unlogip", {
        method: "POST",
        headers: {'Content-Type': 'application/x-www-form-urlencoded'}, 
        body: "confirmunlog=yes"
      }).then(res => {
        switch (res.status) {
            case 200:
                window.location.reload(true) // reload without cache
                break;
            default:
                eID("formreturn").innerHTML = "Error unlogging IP. Please try again later."
                break;
        }
      });
}

eID("unlogFormBtn").onclick = unlogForm
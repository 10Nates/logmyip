function eID(id) { return document.getElementById(id) }

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
//init() replaced internally


// All of this is just for magnification
var magnifier = 1
const docEl = document.documentElement

function decreaseMagnify() {
    magnifier *= 0.9
    if (magnifier < 1) magnifier = 1
    docEl.style.setProperty("--magnify-scale", magnifier)
}

function increaseMagnify() {
    magnifier *= 1.1
    docEl.style.setProperty("--magnify-scale", magnifier)
}

function openBigMap() {
    eID("mapbig").style.setProperty("display", "block")
    eID("magnifyoptions").style.setProperty("display", "block")
}

function closeBigMap() {
    eID("mapbig").style.setProperty("display", "none")
    eID("magnifyoptions").style.setProperty("display", "none")
}

document.addEventListener('DOMContentLoaded', () => {
    eID("decreasemag").onclick = decreaseMagnify
    eID("increasemag").onclick = increaseMagnify
    eID("magnify").onclick= openBigMap
    eID("mapbig").onclick = closeBigMap
})
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
let docWidth = docEl.scrollWidth
let docHeight = docEl.scrollHeight
let scrollXP = .5
let scrollYP = .5

function setScrollPercent() {
    // scroll percentage based on center of screen, i.e. 50% when fully zoomed out
    scrollXP = (window.scrollX + (window.innerWidth / 2)) / docWidth;
    scrollYP = (window.scrollY + (window.innerHeight / 2)) / docHeight;
}

function scrollViewport() {
    let diffW = docEl.scrollWidth - docWidth
    let diffH = docEl.scrollHeight - docHeight
    docWidth = newWidth
    docHeight = newHeight

    window.scroll(window.scrollX + diffW * scrollXP, window.scrollY + diffH * scrollYP)
}

function decreaseMagnify() {
    magnifier *= 0.9
    if (magnifier < 1) magnifier = 1
    setScrollPercent()
    docEl.style.setProperty("--magnify-scale", magnifier)
    scrollViewport()
}

function increaseMagnify() {
    magnifier *= 1.1
    setScrollPercent()
    docEl.style.setProperty("--magnify-scale", magnifier)
    scrollViewport()
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

document.addEventListener('keydown', (e) => {
    switch (e.key) {
        case "+": // US/INT
        case "=": // US/INT
        case "*": // DIN
        case "~": // DIN
            increaseMagnify();
            break;
        case "-":
        case "_":
            if (e.shiftKey) magnifier = 1;
            decreaseMagnify();
            break;
        case "Escape":
            closeBigMap()
            break;
    }
})


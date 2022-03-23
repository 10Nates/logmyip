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
document.addEventListener("showToast", function(event) {
    var snackbar = document.getElementById("success-snackbar");
    if (event.detail.value === "failure") {
        snackbar = document.getElementById("failure-snackbar");
    }

    snackbar.style.visibility = "visible";
    setTimeout(function () { snackbar.style.visibility = "hidden"; }, 3000);
});
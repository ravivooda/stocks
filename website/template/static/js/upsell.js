window.onload = function () {
    if (localStorage.getItem("hasCodeRunBefore") === null) {
        localStorage.setItem("hasCodeRunBefore", 'once');
    } else if (localStorage.getItem("hasCodeRunBefore") === 'once') {
        setTimeout(() => {
            $('#myModal').modal({
                keyboard: false
            })
        }, 1000)
        localStorage.setItem("hasCodeRunBefore", 'twice');
    } else if (!(localStorage.getItem("hasCodeRunBefore") === 'twice')) {
        localStorage.setItem("hasCodeRunBefore", 'once');
    }
}
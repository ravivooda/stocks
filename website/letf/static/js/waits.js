function waitForElm(selector) {
    return new Promise(resolve => {
        if (document.querySelector(selector)) {
            return resolve(document.querySelector(selector));
        }

        const observer = new MutationObserver(mutations => {
            if (document.querySelector(selector)) {
                resolve(document.querySelector(selector));
                observer.disconnect();
            }
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    });
}

function leverageConfigForMobile(elem) {
    waitForElm(elem + '_length').then((elm) => {
        elm.classList.add("d-none")
        elm.classList.add("d-lg-block")
    })
    waitForElm(elem + '_filter').then((elm) => {
        elm.classList.add("d-none")
        elm.classList.add("d-lg-block")
    })
    waitForElm(elem + '_info').then((elm) => {
        elm.classList.add("d-none")
        elm.classList.add("d-lg-block")
    })
}
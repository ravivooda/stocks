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

const elems_for_hiding_on_mobile = ["1x", "1.5x", "1.25x", "1.75x", "2x", "3x", "4x", "-0.5x", "-1x", "-1.25x", "-1.5x", "-2x", "-3x", "false", "no", "variable", "other",];

function setupMobileConfig() {
    for (let i = elems_for_hiding_on_mobile.length - 1; i >= 0; i--) {
        let elem = elems_for_hiding_on_mobile[i];
        waitForElm('#row-select-leverage-' + elem + '_length').then((elm) => {
            elm.classList.add("d-none")
            elm.classList.add("d-lg-block")
        })
        waitForElm('#row-select-leverage-' + elem + '_filter').then((elm) => {
            elm.classList.add("d-none")
            elm.classList.add("d-lg-block")
        })
        waitForElm('#row-select-leverage-' + elem + '_info').then((elm) => {
            elm.classList.add("d-none")
            elm.classList.add("d-lg-block")
        })
    }
}
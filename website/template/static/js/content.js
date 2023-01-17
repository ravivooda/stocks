const contentMap = {
    "leverage_2x_description": "ETFs under Leverage 2x seek daily investment results, before fees and expenses, of 200% of the performance of the aggregated underlying securities. There is no guarantee the fund will meet its stated investment objective."
};

function renderTextForKey(x, id) {
    if (x in contentMap) {
        document.getElementById(id).innerText = contentMap[x];
    } else {
        console.warn("did not find key in content map", x);
        document.getElementById(id).setVisible(false);
    }
}
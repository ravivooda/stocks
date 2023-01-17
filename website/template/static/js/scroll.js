function setupCardScrollForLeverage(leverage) {
    $('#test-row-'+leverage).css('cursor', 'pointer');
    $('#test-row-'+leverage).click(function () {
        $('html, body').animate({
            scrollTop: $("#leverage-"+leverage).offset().top
        }, 1000);
        return false;
    });
}
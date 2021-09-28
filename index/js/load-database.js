(function () {
    const el = $('#select-server-target');

    $.get('database/database.json', function (response) {
        $.each(response, function (index, value) {
            el.append(`<option value='${value.ip}'>${value.name}</option>`);
        });
    });

})();
var target, use_tls, editor;

$('#get-services').click(function(){
    var t = get_valid_target();
    if (target != t) {
        target = t;
        use_tls = $('#use-tls').is(":checked");
    } else {
        return false;
    }
    
    $('.other-elem').hide();
    var button = $(this).html();
    $.ajax({
        url: "server/"+target+"/services",
        global: true,
        method: "GET",
        success: function(res){
            if (res.error) {
                target = "";
                use_tls = "";
                alert(res.error);
                return;
            }
            $("#select-service").html(new Option("Choose Service", ""));
            $.each(res.data, (_, item) => $("#select-service").append(new Option(item, item)));
            $('#choose-service').show();
        },
        error: function(_, _, errorThrown) {
            target = "";
            use_tls = "";
            alert(errorThrown);
        },
        beforeSend: function(xhr){
            xhr.setRequestHeader('use_tls', use_tls);
            $(this).html("Loading...");
        },
        complete: function(){
            $(this).html(button);
        }
    });
});

$('#select-service').change(function(){
    var selected = $(this).val();
    if (selected == "") {
        return false;
    }

    $('#body-request').hide();
    $('#response').hide();
    $.ajax({
        url: "server/"+target+"/service/"+selected+"/functions",
        global: true,
        method: "GET",
        success: function(res){
            if (res.error) {
                alert(res.error);
                return;
            }
            $("#select-function").html(new Option("Choose Method", ""));
            $.each(res.data, (_, item) => $("#select-function").append(new Option(item.substr(selected.length) , item)));
            $('#choose-function').show();
        },
        error: err,
        beforeSend: function(xhr){
            xhr.setRequestHeader('use_tls', use_tls);
            show_loading();
        },
        complete: function(){
            hide_loading();
        }
    });
});

$('#select-function').change(function(){
    var selected = $(this).val();
    if (selected == "") {
        return false;
    }

    $('#response').hide();
    $.ajax({
        url: "server/"+target+"/function/"+selected+"/describe",
        global: true,
        method: "GET",
        success: function(res){
            if (res.error) {
                alert(res.error);
                return;
            }

            generate_editor(res.data.template);
            $("#schema-proto").html(PR.prettyPrintOne(res.data.schema));
            $('#body-request').show();
        },
        error: err,
        beforeSend: function(xhr){
            xhr.setRequestHeader('use_tls', use_tls);
            show_loading();
        },
        complete: function(){
            hide_loading();
        }
    });
});

$('#invoke-func').click(function(){
    var func = $('#select-function').val();
    if (func == "") {
        return false;
    }
    var body = editor.getValue();
    var button = $(this).html();
    $.ajax({
        url: "server/"+target+"/function/"+func+"/invoke",
        global: true,
        method: "POST",
        data: body,
        dataType: "json",
        success: function(res){
            if (res.error) {
                alert(res.error);
                return;
            }
            $("#json-response").html(PR.prettyPrintOne(res.data.result));
            $("#timer-resp span").html(res.data.timer);
            $('#response').show();
        },
        error: err,
        beforeSend: function(xhr){
            xhr.setRequestHeader('use_tls', use_tls);
            $(this).html("Loading...");
        },
        complete: function(){
            $(this).html(button);
        }
    });
});

function generate_editor(content) {
    if(editor) {
        editor.setValue(content);
        return true;
    }
    $("#editor").html(content);
    editor = ace.edit("editor");
    editor.setOptions({
        maxLines: Infinity
    });
    editor.renderer.setScrollMargin(10, 10, 10, 10);
    editor.setTheme("ace/theme/github");
    editor.session.setMode("ace/mode/json");
    editor.renderer.setShowGutter(false);
}

function get_valid_target() {
    t = $('#server-target').val().trim();
    if (t == "") {
        return target;
    }

    ts = t.split("://");
    if (ts.length > 1) {
        $('#server-target').val(ts[1]);
        return ts[1];
    }
    return ts[0];
}

function err(_, _, errorThrown) {
    alert(errorThrown);
}

function show_loading() {
    $('.loading-overlay').show();
}

function hide_loading() {
    $('.loading-overlay').hide();
}
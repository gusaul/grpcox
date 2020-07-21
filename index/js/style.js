var target, use_tls, editor;

window.onload = function () {
    let cachedTarget = localStorage.getItem("server_target")

    console.log(cachedTarget !== "undefined")

    if (cachedTarget !== "undefined") {
        document.getElementById('server-target').value = cachedTarget;
    }
}

$('#get-services').click(function(){
    var t = get_valid_target();

    localStorage.setItem("server_target", t)

    use_tls = "false";
    var restart = "0"
    if($('#restart-conn').is(":checked")) {
        restart = "1"
    }

    // determine whether the proto connection will use local proto or not
    const use_proto = $('#local-proto').is(":checked");

    if (target != t || restart == "1" || use_proto) {
        target = t;
    } else {
        return false;
    }

    // prepare ajax options beforehand
    // makes it easier for local proto to modify some of its properties
    const ajaxProps = {
        url: "server/"+target+"/services?restart="+restart,
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
            $('#choose-service').hide();
            xhr.setRequestHeader('use_tls', use_tls);
            $(this).html("Loading...");
            show_loading();
        },
        complete: function(){
            applyConnCount();
            $(this).html(button);
            hide_loading();
        }
    };

    // modify ajax options if use local proto
    if (use_proto) {
        ajaxProps.method = "POST";
        ajaxProps.enctype = "multipart/form-data";
        ajaxProps.processData = false;
        ajaxProps.contentType = false;
        ajaxProps.cache = false;
        ajaxProps.data = getProtos();
    }

    $('.other-elem').hide();
    var button = $(this).html();
    $.ajax(ajaxProps);
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
            $('#choose-function').hide();
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
            $('#body-request').hide();
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
            $('#response').hide();
            xhr.setRequestHeader('use_tls', use_tls);
            $(this).html("Loading...");
            show_loading();
        },
        complete: function(){
            $(this).html(button);
            hide_loading();
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
    $('.spinner').show();
}

function hide_loading() {
    $('.spinner').hide();
}

$(".connections ul").on("click", "i", function(){
    $icon = $(this);
    $parent = $(this).parent("li");
    var ip = $(this).siblings("span").text();

    $.ajax({
        url: "active/close/" + ip,
        global: true,
        method: "DELETE",
        success: function(res){
            $('[data-toggle="tooltip"]').tooltip('hide');
            if(res.data.success) {
                $parent.remove();
                updateCountNum();
            }
        },
        error: err,
        beforeSend: function(xhr){
            $icon.attr('class', 'fa fa-spinner');
        },
    });
});

function updateCountNum() {
    $(".connections .title span").html($(".connections ul li").length);
}

function applyConnCount() {
    $('[data-toggle="tooltip"]').tooltip('hide');

    $.ajax({
        url: "active/get",
        global: true,
        method: "GET",
        success: function(res){
            $(".connections .title span").html(res.data.length);
            $(".connections .nav").html("");
            res.data.forEach(function(item){
                $list = $("#conn-list-template").clone();
                $list.find(".ip").html(item);
                $(".connections .nav").append($list.html());
            });
            refreshToolTip();
        },
        error: function (_, _, thrownError) {
            console.warn("Failed to update active connections", thrownError)
        },
    });
}

function refreshConnCount() {
    applyConnCount();
    setTimeout(refreshConnCount, 10000);
}

function refreshToolTip() {
    $(function () {
        $('[data-toggle="tooltip"]').tooltip('dispose');
        $('[data-toggle="tooltip"]').tooltip();
    })
}

$(document).ready(function(){
    refreshConnCount();
});

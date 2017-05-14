const cookie_name = "JAMPY_USER_ID";
var songnow = -1;
var vol_val;
var on_repeat = false;
var on_shuffle = false;
var shuffled_songs = new Set();
function showInp(img) {
    let i = $(img);
    i.hide();
    i.parents().children("input").show();
    i.parents().children("input").focus();
}

function addTag(txt, id) {
    console.log("ID:", id);
    $.post("/api/files/" + id + "/" + txt.value, $.cookie(cookie_name));
    txt.value = "";
    window.location.reload()
}

function delTag(tag, id) {
    console.log("del...");
    $.ajax({
        url: "/api/files/" + id + "/" + tag,
        method: 'DELETE',
        data: $.cookie(cookie_name)
    });
    window.location.reload()}

function playSong(name, id, i) {
    var song = $("#song-"+songnow);
    song.removeClass("playing");
    song.addClass("not-playing");
    songnow = i;
    song = $("#song-" +i);
    song.removeClass("not-playing");
    song.addClass("playing");
    $("#time").html("00:00");
    $("#track").html(name);
    $("#pause").show();
    $("#play").hide();
    var w = document.getElementById("audio");
    w.src = "/file/"+id;
    w.play();
    w.volume = $("#vol").val()/100;
}



function search(inp) {
    let txt = inp.value;
    if(txt === "") {
        window.location.assign('/');
        return
    }
    window.location.assign("/search?" +
        (txt[0] === '#' ? "tag="+txt.substring(1) : "name="+txt)
    );
}

$(document).ready(() => {
    // makePage();
    $("#logout").click(() => {
        $.cookie(cookie_name, "", {
            expires: -1
        });
        window.location.reload()
    });
    var song = document.getElementById("audio");
    song.addEventListener('timeupdate',function (){
        let curtime = song.currentTime;
        let s = parseInt(curtime % 60);
        let m = parseInt((curtime / 60) % 60);
        s = (s >= 10) ? s : "0" + s;
        m = (m >= 10) ? m : "0" + m;
        $("#time").html(m + ':' + s );
    });
    song.addEventListener("ended", function () {
        $("#pause").hide();
        $("#play").show();
        $("#time").html("00:00");
        $("#forward-img").trigger("onclick");
    });
    $('#vol').change(function () {
        document.getElementById('audio').volume = $(this).val()/100;
    });
    $("#vol-img").click(function () {
        var vol = $("#vol");
        if (vol.val() == 0) {
            vol.val(vol_val);
            vol_val = 0;
            document.getElementById('audio').volume = vol.val()/100;
            $(this).attr("src", "./static/sound-w.png");
        } else {
            vol_val = vol.val();
            vol.val(0);
            document.getElementById('audio').volume = 0;
            $(this).attr("src", "./static/no-sound-w.png");
        }
    });
    $("#forward-img").click(function () {
        var next;
        if (on_shuffle === true) {
            if (shuffled_songs.size === num_songs) {
                shuffled_songs.clear();
            }
            next = Math.floor(Math.random()*(num_songs-1));
            while (shuffled_songs.has(next)) {
                next = Math.floor(Math.random()*(num_songs-1));
            }
            shuffled_songs.add(songnow);
        } else {
            next = songnow+1;
            if (next === num_songs) {
                next = 0;
            }
        }
        console.log(songnow, next);
        $("#song-"+next).children(".play-button").trigger("onclick");
    });
    $("#backwards-img").click(function () {
        var next = songnow - 1;
        if (next === -1) {
            next = num_songs -1;
        }
        $("#song-"+next).children(".play-button").trigger("onclick");
    });
    $("#repeat-img").click(function () {
        if (on_repeat === false) {
            on_repeat = true;
            $(this).removeClass("not-active-img");
            $(this).addClass("active-img");
        } else {
            on_repeat = false;
            $(this).removeClass("active-img");
            $(this).addClass("not-active-img");
        }
    });
    $("#shuffle-img").click(function () {
        if (on_shuffle === false) {
            on_shuffle = true;
            $(this).removeClass("not-active-img");
            $(this).addClass("active-img");
        } else {
            on_shuffle = false;
            $(this).removeClass("active-img");
            $(this).addClass("not-active-img");
        }
    })
});

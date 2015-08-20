var curuser = "";
var curlab = "";

$(function(){
  // forms
  $("button#groupsubmit").click(function(){
    $("form#groupselection").submit();
  });

  $("#rebuild").click(function(event){
    var lab = curlab;
    var user = curuser;
    $("div.alert").show(200);
    $("div.alert").removeClass("alert-primary alert-danger alert-success").addClass("alert-warning").text("Running build");
    $.post("/event/manualbuild", {"course": course, "user": user, "lab": lab}, function(){
      $("div.alert").removeClass("alert-warning").addClass("alert-success").text("Successfull rebuild. Build log updated.");
      loadLabResult(user, lab);
    }).fail(function(){
      $("div.alert").removeClass("alert-warning").addClass("alert-danger").text("Rebuild failure");
    });
    event.preventDefault();
    return false
  });

  // sec nav bar links
  $('a.indvlabtab').click(function (e) {
    var lab = $(this).attr('lab');
    loadLabResult(username, lab);
    $('div.result-content').hide();
    $('div#resultview').show();

    curuser = username;
    curlab = lab;

    // nav active marking
    $('a.list-group-item').removeClass("active");
    $(this).addClass("active");
    $("div.alert").hide();
  });

  $('a#groupinfotab').click(function(){
    $('div.result-content').hide();
    $('div#groupsummaryview').show();

    // nav active marking
    $('a.list-group-item').removeClass("active");
    $(this).addClass("active");
  });

  $('a#groupregtab').click(function(){
    $('div.result-content').hide();
    $('div#groupsignupview').show();

    // nav active marking
    $('a.list-group-item').removeClass("active");
    $(this).addClass("active");
  });

  $('a.reviewpubtab').click(function(){
    $('div.result-content').hide();
    $('div#publishreviewview').show();

    // nav active marking
    $('a.list-group-item').removeClass("active");
    $('.reviewpubtab').addClass("active");
  });

  $('a.indvlabtab:eq(' + labnum + ')').click();

  $("button#random").click(function(){
    $.post("/course/requestrandomgroup", {"course": course}, function(){
      alert("You have been added to the list over people who wants random groups assignment.");
    });
  });
});

var curuser = "";
var curlab = "";

var loadLabResult = function(user, lab){
  $('span#lab-headline').text(lab);
  $.getJSON("/course/ciresutls", {"Labname": lab, "Course": course, "Username": user}, function(data){
  // updates text fields
    $("#status").text("Status: ").append(data.Status);
    $("#passes").text("Number of passed tests: ").append(data.NumPasses);
    $("#fails").text("Number of failed tests: ").append(data.NumFails);
        
    // updates code
    $("code#logs").text("");
    data.Log.forEach(function(t, i) {
      $("code#logs").append(" # ").append($(document.createTextNode(t))).append("<br>");
    });

    // updates build and push times. 
    var d = new Date(data.Timestamp);
    $("p#timedate").text("Build time: ").append(d.toLocaleDateString() + " - " + d.toLocaleTimeString());
    var d2 = new Date(data.PushTime);
    $("#pushtime").text("Code delievered: ").append(d2.toLocaleDateString() + " - " + d2.toLocaleTimeString());

    // updates processbar
    var pbar = $("div.progress > div.progress-bar");
    pbar.removeClass("progress-bar-success progress-bar-warning progress-bar-danger progress-bar-striped");
    if(data.NumBuildFailure == 0) {
      pbar.text(data.TotalScore+"%").attr("aria-valuenow", data.TotalScore).css("width", data.TotalScore + "%");
      if(data.TotalScore < 40 && data.TotalScore >= 6){
        pbar.addClass("progress-bar-danger");
      } else if(data.TotalScore < 6) {
        pbar.addClass("progress-bar-danger");
        pbar.attr("aria-valuenow", 6).css("width", "6%")
      } else if(data.TotalScore >= 40 && data.TotalScore < 60){
        pbar.addClass("progress-bar-warning");
      }
    } else {
      pbar.text("Build Failure!");
      pbar.attr("aria-valuenow", 100).css("width", "100%");
      pbar.addClass("progress-bar-danger progress-bar-striped");
    }

    // update test table
    $("table#testresultlist > tfoot > tr > .totalscore").text(data.TotalScore+"%");
    var testtable = $("table#testresultlist > tbody");
    testtable.text("");
    if(data.TestScores != null){
      data.TestScores.forEach(function(data, i){
        testtable.append("<tr><td>" + (i + 1) + "</td><td>" + data.TestName + "</td><td>" + data.Score + "/" + data.MaxScore + " pts</td><td>" + data.Weight + " pts</td></tr>\n");
      });
    }
  }).fail(function(){
    $("#status").text("Status: Nothing built yet.");
    $("#passes").text("Number of passed tests: -");
    $("#fails").text("Number of failed tests: -");
    $("p#timedate").text("Build time: -");
    $("#pushtime").text("Code delievered: -");
    $("div.progress > div.progress-bar").removeClass("progress-bar-success progress-bar-warning progress-bar-danger progress-bar-striped").attr("aria-valuenow", 6).css("width", "6%").text("0%");
    $("code#logs").text("# There is no build for this lab yet.");
    $("table#testresultlist > tfoot > tr > .totalscore").text("0%");
    $("table#testresultlist > tbody").html("");
  });
}

var tablerowlink = function(href, target){
  window.open(href, target);
}

$(function(){
  // forms
  $('form#publishreviewform').submit(function(){
    // TODO: validate
    $.post("/review/publish", $(this).serialize(), function(base){
      var data = jQuery.parseJSON(base);
      a = $('#publishreviewview > .alert');
      a.removeClass("alert-success alert-danger");
      if(data.Error) {
        a.addClass("alert-danger");
        a.text("Could not publish code review. Message: " + data.Errormsg);
      } else {
        a.addClass("alert-success");
        a.html('Code Review published! <a href="' + data.CommitURL + '" target="_blank">Take a look at it.<a/>');
      }
      a.show();
    });
    event.preventDefault();
    return false
  });

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

  $('a#reviewlisttab').click(function(){
    $('div.result-content').hide();
    $('div#listreviewsview').show();

    // nav active marking
    $('a.list-group-item').removeClass("active")
    $(this).addClass("active")

    $.getJSON("/review/list", {"course": course}, function(data){
      def = $('#reviewlisttable > tbody > tr').last();
      $('#reviewlisttable > tbody').html(def);
      data.Reviews.forEach(function(r, i) {
        if (r.Desc.length > 75) {
          desc = r.Desc.substring(0, 75) + "...";
        } else {
          desc = r.Desc;
        }
        $('#reviewlisttable > tbody > tr').first().before("<tr onclick=\"tablerowlink('" + r.URL + "', '_blank')\"><td>" + (i + 1) + "</td><td>" + r.Title + "</td><td><i>" + desc + "</i></td><td><a href=\"#\">Go to review</a></td></tr>");
      });
    });
  });

  $('a.indvlabtab:eq(' + labnum + ')').click();

  // allow use of tabs in textareas
  $("textarea").keydown(function(e) {
    if(e.keyCode === 9) { // tab was pressed
      // get caret position/selection
      var start = this.selectionStart;
      var end = this.selectionEnd;

      var $this = $(this);
      var value = $this.val();

      // set textarea value to: text before caret + tab + text after caret
      $this.val(value.substring(0, start) + "    " + value.substring(end));

      // put caret at right position again (add one for the tab)
      this.selectionStart = this.selectionEnd = start + 4;

      // prevent the focus lose
      e.preventDefault();
    }
  });
});
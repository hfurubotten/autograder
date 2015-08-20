// loads lab results from server and updates html.
var loadLabResult = function(user, lab){
  $('span#lab-headline').text(lab);
  $.getJSON("/course/ciresutls", {"Labname": lab, "Course": course, "Username": user}, function(data){
    // updates text fields
    $("#status").text("Status: ").append(data.Status);
    $("#passes").text("Number of passed tests: ").append(data.NumPasses);
    $("#fails").text("Number of failed tests: ").append(data.NumFails);
    $("#buildid").text("Build ID: ").append(data.ID);

    // updates code
    $("code#logs").text("");
    data.Log.forEach(function(t, i) {
        $("code#logs").append((i + 1) + ": ").append($(document.createTextNode(t))).append("<br>");
    });

    // updates build and push times.
    var timeformat = "DD/MM/YYYY HH:mm:ss"
    var buildtime = moment(data.Timestamp).format(timeformat);
    $("p#timedate").text("Build date: ").append(buildtime);
    var pushtime = moment(data.PushTime).format(timeformat);
    $("#pushtime").text("Code delievered: ").append(pushtime);
    var buildduration = moment(data.BuildTime/(1000*1000), "x").format("mm [min] ss.SS [sec]");
    $("#buildtime").text("Execution time: ").append(buildduration);

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
    clearLabResults();
  });
}

var clearLabResults = function() {
  $("#status").text("Status: Nothing built yet.");
  $("#passes").text("Number of passed tests: -");
  $("#fails").text("Number of failed tests: -");
  $("p#timedate").text("Build date: -");
  $("#buildtime").text("Execution time: -");
  $("#buildid").text("Build ID: -");
  $("#pushtime").text("Code delievered: -");
  $("div.progress > div.progress-bar").removeClass("progress-bar-success progress-bar-warning progress-bar-danger progress-bar-striped").attr("aria-valuenow", 6).css("width", "6%").text("0%");
  $("code#logs").text("# There is no build for this lab yet.");
  $("table#testresultlist > tfoot > tr > .totalscore").text("0%");
  $("table#testresultlist > tbody").html("");
}

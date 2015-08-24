$("#rebuild").click(function(event){
  var lab = curlab;
  var user = username;
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

$("#approve").click(function(){
  if(confirm("Are you sure you want to approve this lab?")){
    $.post("/course/approvelab", {"Course": course, "User": username, "Approve": true, "Labnum": curlabnum}, function(){
      loadLabResult(username, curlab);
    });
  }
  event.preventDefault();
  return false
});

$(".labtab").click(function(){
  $('a.list-group-item').removeClass('active');
  $(this).addClass("active");
  curlab = $(this).attr("lab");
  curlabnum = $(this).attr("labnum");

  loadLabResult(username, curlab);
  loadNotes()
});

$(".summary").click(function(){
  $('a.list-group-item').removeClass('active');
  $(".summary").addClass("active");
  updatesummary();
});


$(window).on('hashchange', function(){
  $('section').hide();
  $(location.hash).show();
  //$('a.list-group-item').removeClass('active');
  //$('a[href=' + location.hash + ']').addClass('active');
});

var loadNotes = function(){
  $.getJSON("/course/notes",
    {"Course": course, "Username": username, "Group": groupid, "Labnum": curlabnum},
    function(data){
      $(".labnotes").val(data.Notes);
      $("button#notessubmit").removeClass("btn-primary").addClass("btn-success")
    }).fail(function(){
      $(".labnotes").val("");
    });
}

$("form#notes").submit(function(event){
  var notes = $("textarea[name=Notes]").val();
  $.post("/course/notes",
    {"Course": course, "Username": username, "Group": groupid, "Labnum": curlabnum, "Notes": notes},
    function(){
      loadLabResult(username, curlab);
  });

  event.preventDefault();
  return false
});

$("textarea[name=Notes]").focus(function(){
  $("button#notessubmit").removeClass("btn-success").addClass("btn-primary")
})

var addtosummarypanel = function(labname, status, score, notes, tablebody){
  $("#summarypanel").append("<div class=\"panel panel-default\">" +
    "<div class=\"panel-heading\">" +
      "<h3 class=\"panel-title\"><strong>"+labname+"</strong> <span class=\"pull-right\">Score: "+score+"% | Status: "+status+"</span></h3>"+
    "</div>" +
    "<div class=\"panel-body\">"+
      "<p><strong>Notes:</strong></b></p>"+
      "<p>"+notes+"</p>"+

      "<hr>"+

      "<table id=\"testresultlist\" class=\"table table-striped\">"+
        "<thead>"+
          "<tr>"+
            "<th>#</th>"+
            "<th>Test name</th>"+
            "<th>Score</th>"+
            "<th>Weight</th>"+
          "</tr>"+
        "</thead>"+
        "<tbody>"+
          tablebody +
        "</tbody>"+
        "<tfoot>"+
          "<tr>"+
            "<td></td>"+
            "<td>Total score:</td>"+
            "<td class=\"totalscore\">"+score+"%</td>"+
            "<td>100%</td>"+
          "</tr>"+
        "</tfoot>"+
      "</table>"+
    "</div>"+
  "</div>");
}

var updatesummary = function(){
  $.getJSON("/course/cisummary", {"Course": course, "Username": username}, function(data){
    $("#summarypanel").text("");
    $.each(data.Summary, function(labname, s){
      if(labname == "") {
        return
      }

      notes = data.Notes[labname]

      // update test table
      tablebody = ""
      if(s.TestScores != null){
        s.TestScores.forEach(function(data, i){
          tablebody = tablebody + "<tr><td>" + (i + 1) + "</td><td>" + data.TestName + "</td><td>" + data.Score + "/" + data.MaxScore + " pts</td><td>" + data.Weight + " pts</td></tr>\n";
        });
      }

      addtosummarypanel(labname, s.Status, s.TotalScore, notes, tablebody)
    });
  });
}

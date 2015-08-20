// When the document is ready it will activate the student tab and update the summary for each of the students and groups.
$(function(){
  $('.date').datetimepicker({
    //format: "YYYY-MM-DD HH:mm:ss.SSSSSSSSS ZZ zz",
    format: "DD/MM/YYYY HH:mm",
    extraFormats: [ 'YYYY-MM-DD HH:mm:ss.SSSSSSSSS' ]
  });

  var page = location.hash.trim();
  if(page){
    $('section').hide();
    $(page).show();
    $('a.list-group-item').removeClass('active');
    $('a[href=' + page + ']').addClass('active');
  } else {
    $('section').hide();
    if(labtype == 1){
      $('#groupresults').show();
      $('a[href=#groupresults]').addClass('active');
    } else {
      $('#studentresults').show();
      $('a[href=#studentresults]').addClass('active');
    }
  }

  $(window).on('hashchange', function(){
    $('section').hide();
    $(location.hash).show();
    $('a.list-group-item').removeClass('active');
    $('a[href=' + location.hash + ']').addClass('active');
  });

  $("table#studentresults > tbody > tr").each(updatesummary);
  $("table#groupresults > tbody > tr").each(updatesummary);
});

// updatesummary is a finction that will load a summary for a course and update the student or group list.
var updatesummary = function(index, element){
  var username = $(element).attr("id");
  $.getJSON("/course/cisummary", {"Course": course, "Username": username}, function(data){
    $.each(data.Summary, function(labname, s){
      if(labname == "") {
        return
      }
      
      $("tr#" + data.User + " > td." + labname).text(s.TotalScore + "%")
      if(s.Status.toLowerCase() == "Active lab assignment".toLowerCase()){
        $("tr#" + data.User + " > td." + labname).addClass("info");
      } else if(s.Status.toLowerCase() == "Approved".toLowerCase()){
        $("tr#" + data.User + " > td." + labname).addClass("success");
      }
    });
  }).fail(function(){

  });
}

// addassistant will send a request to the server for adding a student to the list of teaching assistants.
function addassistant(username, remove){
  if(confirm("Are you sure you want to make " + username + " a teaching assistant?")){
    $.post("/course/addassistant", {"course": course, "assistant": username}, function(base){
      if(remove) {
        $("tr#" + username).slideUp(500);
      }
    }).fail(function(){
      $("#pendinguseralert > span.inputtext").text("Error adding the user as assistant.");
      $("div#pendinguseralert").show();
    });
  }
}

// addassistant will send a request to the server for adding a student to the list of teaching assistants.
function removeassistant(username, remove){
  if(confirm("Are you sure you want to remove " + username + " from the teaching staff?")){
    $.post("/course/removeassistant", {"course": course, "assistant": username}, function(base){
      if(remove) {
        $("tr#" + username).slideUp(500);
      }
    }).fail(function(){
      $("#pendinguseralert > span.inputtext").text("Error adding the user as assistant.");
      $("div#pendinguseralert").show();
    });
  }
}

// removependinguser will send a request to the server and remove a pending user.
function removependinguser(username){
  $.post("/course/removepending", {"course": course, "user": username}, function(base){
      $("tr#" + username).slideUp(500);
    }).fail(function(){
      $("#pendinguseralert > span.inputtext").text("Error removing the user from the list.");
      $("div#pendinguseralert").show();
    });
}

// approveuser will send a request to the server to approve a user.
function approveuser(username) {
  $.post("/course/approvemember/" + course, {"user": username}, function(base){
    var data = jQuery.parseJSON(base);
    if (!data.Error) {
      $("tr#" + username).slideUp(500);
    } else {
      $("#pendinguseralert > span.inputtext").text(data.ErrorMsg);
      $("div#pendinguseralert").show();
    }
  });
}

// removeuserfromcourse will send a request to the server for removing a student from a course.
function removeuserfromcourse(username, remove){
  if(confirm("Are you sure you want to remove " + username + " from the course?")){
    $.post("/course/removemember", {"course": course, "user": username}, function(base){
      if(remove) {
        $("tr#" + username).slideUp(500);
      }
    }).fail(function(){
      $("#pendinguseralert > span.inputtext").text("Error adding the user as assistant.");
      $("div#pendinguseralert").show();
    });
  }
}

// approvegroup will send a request to the server to approve a group.
function approvegroup(groupid) {
  $.post("/course/approvegroup", {"course": course, "groupid": groupid}, function(base){
    var data = jQuery.parseJSON(base);
    if (!data.Error) {
      $("tr#group" + groupid).slideUp(500);
    } else {
      $("#pendinggroupalert > span.inputtext").text(data.ErrorMsg);
      $("div#pendinggroupalert").show();
    }
  });
}

// removegroup will send a request to the server to remove a group.
function removegroup(groupid){
  if(confirm("Are you sure you want to remove this group?")){

    if (typeof groupid == "string") {
      groupid = groupid.substring(5)
    }

    $.post("/course/removegroup", {"course": course, "groupid": groupid}, function(base){
      $("tr#group" + groupid).slideUp(500);
    }).fail(function(){
      $("#pendinggroupalert > span.inputtext").text("Error removing the group from the list.");
      $("div#pendinggroupalert").show();
    });
  }
}

// addtoexistinggroup will send a request to the server to add additional members to a existing group.
function addtoexistinggroup(groupid) {
  $('form#groupselection > input[name=groupid]').val(groupid);
  var input = $('form#groupselection').serialize()

  $.post("/group/addmember", input, function(base){
    var data = jQuery.parseJSON(base);
    if(data.Error){
      $('#pendinggroupalert').text("Error: " + data.ErrorMsg);
    } else if(data.Added){
      location.reload(true);
    } else {
      $('#pendinggroupalert').text("Unknown error.");
    }
  }).fail(function(){
    $('#pendinggroupalert').text("Communication error.");
  });
}

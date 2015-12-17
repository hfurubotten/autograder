$("#testplagiarism").click(function(event){
  return performApTestClick("individual");
});

$("#grouptestplagiarism").click(function(event){
  return performApTestClick("group");
});

function performApTestClick(labs) {
  $.post("/event/apmanualtest", {"course": course, "labs": labs}, function(){
    alert("The anti-plagiarism command was sent. It will take several minutes at the minimum to process. Please be patient.");
  }).fail(function(){
    alert("The anti-plagiarism command failed.");
  });
  event.preventDefault();
  return false
}

// loads anti-plagiarism results for a user's lab from server and updates html.
var loadLabApResults = function(user, lab){
  $.getJSON("/course/aplabresults", {"Labname": lab, "Course": course, "Username": user}, function(data){
    // Moss results and link
    if (data.MossPct == 0.0) {
    	$("#mossResults").text("0%");
    	$("#mossBtn").hide();
    }
    else {
    	$("#mossResults").text(data.MossPct).append("%");
    	$("#mossBtn").show();
    	var tmp = 'showApDetails("'
    	var onClickCmd = tmp.concat(data.MossURL, '")');
    	$("#mossBtn").attr("onclick", onClickCmd);
    }
    // JPlag results and link
    if (data.JplagPct == 0.0) {
    	$("#jplagResults").text("0%");
    	$("#jplagBtn").hide();
    }
    else {
    	$("#jplagResults").text(data.JplagPct).append("%");
    	$("#jplagBtn").show();
    	var tmp = 'showApDetails("'
    	var onClickCmd = tmp.concat(data.JplagURL, '")');
    	$("#jplagBtn").attr("onclick", onClickCmd);
    }
    // dupl results and link
    if (data.DuplPct == 0.0) {
    	$("#duplResults").text("False");
    	$("#duplBtn").hide();
    }
    else {
    	$("#duplResults").text("True");
    	$("#duplBtn").show();
    	var tmp = 'showApDetails("'
    	var onClickCmd = tmp.concat(data.DuplURL, '")');
    	$("#duplBtn").attr("onclick", onClickCmd);
    }
  }).fail(function(){
    $("#mossResults").text("").append("-1% : Error");
    $("#mossBtn").hide();
    $("#jplagResults").text("").append("-1% : Error");
    $("#jplagBtn").hide();
    $("#duplResults").text("").append("-1% : Error");
    $("#duplBtn").hide();
  });
}

// loads anti-plagiarism results for a user's lab from server and updates
// the cell color in the user's table row
var loadUserApResults = function(index, element){
  var username = $(element).attr("id");
  $.getJSON("/course/apuserresults", {"Course": course, "Username": username}, function(data){
    $.each(data, function(labname, s){
      if(labname == "") {
        return
      }
      
      // Count the tools which found plagiarism
      var count = 0;
      if (s.MossPct > 15.0) {
        count++;
      }
      if (s.JplagPct > 15.0) {
        count++;
      }
      if (s.DuplPct > 0.0) { // dupl is either 1 or 0
        count++;
      }

      // Change cell color depending on the number of tools which found plagiarism.
      if (count == 1) {
        $("tr#" + username + " > td." + labname).css('background-color', '#f7bbbb');
      }
      else if (count == 2) {
        $("tr#" + username + " > td." + labname).css('background-color', '#f08080');
      }
      else if (count == 3) {
        $("tr#" + username + " > td." + labname).css('background-color', '#e73232');
      }
    });
  }).fail(function(){

  });
}

// Show the specific anti-plagiarism details in another window.
function showApDetails(url) {
  window.open(url);
  return true
}

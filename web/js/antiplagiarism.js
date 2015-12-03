$("#testplagiarism").click(function(event){
  return performTestPlagiarismClick("individual");
});

$("#grouptestplagiarism").click(function(event){
  return performTestPlagiarismClick("group");
});

function performTestPlagiarismClick(labs) {
  $.post("/event/manualtestplagiarism", {"course": course, "labs": labs}, function(){
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
    	var tmp = "showApDetails("
    	var onClickCmd = tmp.concat(data.MossURL, ")");
    	alert(onClickCmd);
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
    	var tmp = "showApDetails("
    	var onClickCmd = tmp.concat(data.JplagURL, ")");
    	alert(onClickCmd);
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
    	var tmp = "showApDetails("
    	var onClickCmd = tmp.concat(data.DuplURL, ")");
    	alert(onClickCmd);
    	$("#duplBtn").attr("onclick", onClickCmd);
    }
  }).fail(function(){
    $("#mossResults").text("").append("-1% : Error");
    $("#jplagResults").text("").append("-1% : Error");
    $("#duplResults").text("").append("-1% : Error");
  });
}

// loads anti-plagiarism results for a user's lab from server and updates html.
var loadUserApResults = function(index, element){
  var username = $(element).attr("id");
  $.getJSON("/course/apuserresults", {"Course": course, "Username": username}, function(data){
    $.each(data, function(labname, s){
      if(labname == "") {
        return
      }
      var count = 0;
      
      if (s.MossPct > 0.0) {
        count++;
      }
      if (s.JPlagPct > 0.0) {
        count++;
      }
      if (s.DuplPct > 0.0) {
        count++;
      }      	

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
	alert(url);
	//window.open("http://www.google.com/");
}
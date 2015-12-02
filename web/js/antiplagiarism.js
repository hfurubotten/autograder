$("#testplagiarism").click(function(event){
  return performTestPlagiarismClick("individual");
});

$("#grouptestplagiarism").click(function(event){
  return performTestPlagiarismClick("group");
});

function performTestPlagiarismClick(labs) {
  $.post("/event/manualtestplagiarism", {"course": course, "labs": labs}, function(){
    alert("The anti-plagiarism command was sent. It will take several minutes at the minimum to process. Please be patient. The results will appear in x.");
  }).fail(function(){
    alert("The anti-plagiarism command failed.");
  });
  event.preventDefault();
  return false
}

// loads anti-plagiarism results from server and updates html.
var loadApResults = function(user, lab){
  $.getJSON("/course/apresults", {"Labname": lab, "Course": course, "Username": user}, function(data){
    $("#mossResults").text(data.MossPct).append("%");
    $("#jplagResults").text(data.JplagPct).append("%");
    if (data.DuplPct == 0.0) {
    		$("#duplResults").text("False");
    }
    else {
    		$("#duplResults").text("True");
    }
  }).fail(function(){
    $("#mossResults").text("").append("-1% : Error");
    $("#jplagResults").text("").append("-1% : Error");
    $("#duplResults").text("").append("-1% : Error");
  });
}

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
	// updates text fields
	$("#mossResults").text("Status: ").append(data.Status);
	$("#jplagResults").text("Number of passed tests: ").append(data.NumPasses);
	$("#duplResults").text("Number of failed tests: ").append(data.NumFails);
}
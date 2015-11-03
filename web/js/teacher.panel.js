$("#testplagiarism").click(function(event){
  return performTestPlagiarismClick("individual");
});

$("#grouptestplagiarism").click(function(event){
  return performTestPlagiarismClick("group");
});

function performTestPlagiarismClick(labs) {
  $("div.alert").show(200);
  $("div.alert").removeClass("alert-primary alert-danger alert-success").addClass("alert-warning").text("Running anti-plagiarism");
  $.post("/event/manualtestplagiarism", {"course": course, "labs": labs}, function(){
    $("div.alert").removeClass("alert-warning").addClass("alert-success").text("Anti-plagiarism test started.");
  }).fail(function(){
    $("div.alert").removeClass("alert-warning").addClass("alert-danger").text("Anti-plagiarism test failed to start.");
  });
  event.preventDefault();
  return false
}

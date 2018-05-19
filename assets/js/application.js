require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap-sass/assets/javascripts/bootstrap.js");
$(() => {

});

$('.row .btn').on('click', function(e) {
    e.preventDefault();
    var $this = $(this);
    var $collapse = $this.closest('.collapse-group').find('.collapse');
    $collapse.collapse('toggle');
});
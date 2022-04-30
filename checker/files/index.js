function main() {
  const showLink = a => {
    $('a').removeClass('active');
    $(a).addClass('active');
    let host = $(a).attr('data-href');
    $('#iframe').attr('src', 'http://' + host);
    $('#current-host').text(host);
  };
  $('.subdomain').on('click', el => {
    showLink(el.target);
  });
  $('.goog').on('click', el => {
    let host = $(el.target).attr('data-href');
    let q = 'site:' + host;
    let uri = 'https://google.com/search?q=' + encodeURIComponent(q);
    window.open(uri);
  });
  $(document.body).on('keydown', e => {
    const showSibling = delta => {
      let as = Array.from($('a.subdomain'));
      let actives = as.filter(el => el.className.includes('active'));
      if (!actives.length) {
        actives = [as[0]];
      }
      let index = as.indexOf(actives[0]);
      let nextIndex = index + delta;
      if (nextIndex >= 0 && nextIndex < as.length) {
        let a = as[nextIndex];
        showLink(as[nextIndex]);
      }
    }
    // https://stackoverflow.com/questions/4104158/jquery-keypress-left-right-navigation
    if (e.keyCode == 37) { // left
      showSibling(-1);
    } else if (e.keyCode == 39) { // right
      showSibling(+1);
    }
  });
  {
    let first = $('.subdomain')[0];
    showLink(first);
  }
}
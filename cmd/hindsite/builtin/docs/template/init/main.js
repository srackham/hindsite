function setHeaderLink(heading) {
    var id = heading.getAttribute('id');
    if (id) {
        var link = document.createElement('a');
        link.classList.add('header-link');
        link.setAttribute('href', '#' + id);
        heading.appendChild(link);
    }
}

function appendTocEntry(heading) {
    var id = heading.getAttribute('id');
    if (heading.classList.contains('no-auto-toc')) {
        return;
    }
    var container = document.getElementById('auto-toc');
    if (container === null) {
        return;
    }
    var tocLink = document.createElement('a');
    tocLink.setAttribute('href', '#' + id);
    tocLink.textContent = heading.textContent;
    var tocEntry = document.createElement('div');
    tocEntry.setAttribute('class', heading.tagName.toLowerCase());
    tocEntry.appendChild(tocLink);
    container.appendChild(tocEntry);
}

document.onclick = function (event) {
    if (event.target.matches('#toc-button') || event.target.matches('#toc a') && isSmallScreen()) {
        toggleToc();
    }
}

function toggleToc() {
    document.body.classList.toggle('hide-toc');
}

function isSmallScreen() {
    return window.matchMedia('(min-width: 1px) and (max-width: 800px)').matches;
}

// matches() polyfill for old browsers.
if (!Element.prototype.matches) {
    var p = Element.prototype;
    if (p.webkitMatchesSelector) // Chrome <34, SF<7.1, iOS<8
        p.matches = p.webkitMatchesSelector;
    if (p.msMatchesSelector) // IE9/10/11 & Edge
        p.matches = p.msMatchesSelector;
    if (p.mozMatchesSelector) // FF<34
        p.matches = p.mozMatchesSelector;
}
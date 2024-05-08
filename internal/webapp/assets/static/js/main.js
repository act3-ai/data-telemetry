// enable htmx debugging
// htmx.logAll();

// Copied from getbootstrap.com, enables tooltip
var tooltipTriggerList = [].slice.call(
    document.querySelectorAll('[data-bs-toggle="tooltip"]')
);
var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
    return new bootstrap.Tooltip(tooltipTriggerEl);
});

// Copied from getbootstrap.com, enables popover
var popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'))
var popoverList = popoverTriggerList.map(function (popoverTriggerEl) {
    return new bootstrap.Popover(popoverTriggerEl)
})

// Function to copy text
function copyToClipboard(text) {
    if (navigator.clipboard) {
        navigator.clipboard.writeText(text)
    } else {
        // https://stackoverflow.com/questions/51805395/navigator-clipboard-is-undefined
        // The clipboard is only available in Firefox if served over HTTPS (not available in HTTP)
        alert("Failed to copy to clipboard")
    }
}

var selectorIndex = 0;
function addSelector(type, value) {
    var original = document.getElementById(type + 'SelectorDiv');
    // var removeBtn = document.getElementById('removeSelectorButton');

    var clone = original.cloneNode(true); // "deep" clone
    original.getElementsByClassName(type + "SelectorInput")[0].value = ""; // reset original input field
    clone.id = "selectorDiv" + ++selectorIndex;
    clone.className = "col-10 mt-2";

    // var btnClone = removeBtn.cloneNode(true); // "deep" clone
    // btnClone.className = "col-2";

    original.parentNode.appendChild(clone);
    original.closest('form').submit();
    // clone.parentNode.appendChild(btnClone);
}

function removeSelector(type, index) {
    var input = document.getElementById(type + "SelectorDiv" + index);
    var btn = document.getElementById(type + "RemoveSelectorButton" + index);

    var parent = input.parentNode;

    parent.removeChild(btn);
    parent.removeChild(input);

    var form = parent.closest('form');

    // HACK remove empty selectors before submitting
    var allInputs = form.getElementsByTagName('input');
    for (var i = 0; i < allInputs.length; i++) {
      var input = allInputs[i];
      if (input.name && !input.value) {
        console.log("Found", input.name)
        input.name = '';
      }
    }

    form.submit();
}

// allow HTMX swaps on any 400 or greater status code
document.addEventListener("DOMContentLoaded", (event) => {
    document.body.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.xhr.status >= 400) {
            evt.detail.shouldSwap = true;
            evt.detail.isError = false;
        }
    });
})

// Auto re-sizing of textarea based on content - https://stackoverflow.com/questions/454202/creating-a-textarea-with-auto-resize
const tx = document.getElementsByTagName("textarea");
for (let i = 0; i < tx.length; i++) {
  tx[i].setAttribute("style", "height:" + (tx[i].scrollHeight) + "px;overflow-y:hidden;");
  tx[i].addEventListener("input", OnInput, false);
}

function OnInput() {
  this.style.height = "auto";
  this.style.height = (this.scrollHeight) + "px";
}

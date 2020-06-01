// table input for context metadata
const ctxMetadataTable = $('#ctx-metadata-table');

// new row for each context being added
const newTr = `
    <tr>
        <td>
          <span class="table-remove">
            <button type="button" class="btn btn-danger btn-rounded btn-sm my-0">
              <i class="fa fa-times"></i>
            </button>
          </span>
        </td>
        <td class="ctx-metadata-input-field pt-3-half" contenteditable="true"></td>
        <td class="ctx-metadata-input-field pt-3-half" contenteditable="true"></td>
     </tr>`;

// helper variable to contains all of the context metadata input
let ctxArr = [];

// helper variable to contain the usage of metadata
let ctxUse = false;

// ctx metadata event listener
(function () {
    // add event listener on ctx metadata checkbox
    const ctxMetadataSwitch = document.getElementById("ctx-metadata-switch");
    ctxMetadataSwitch.addEventListener("change", function(event) {
        const { checked } = event.target;
        ctxUse = checked;
        toggleDisplayCtxMetadataTable(checked);
    });

    // remove for each row in ctx metadata table
    ctxMetadataTable.on('click', '.table-remove', function () {
        $(this).parents('tr').detach();
    });

    // add new row
    ctxMetadataTable.on('click', '.table-add', () => {
        $('tbody').append(newTr);
    });

    // only allow any paste action with plain text
    ctxMetadataTable.on('paste', '.ctx-metadata-input-field', function (e) {
        // cancel paste
        e.preventDefault();
        // get text representation of clipboard
        const text = (e.originalEvent || e).clipboardData.getData('text/plain');
        // insert text manually
        document.execCommand("insertHTML", false, text);
    });

}());

// toggle ctx metadata display
// will show the key-value pairs table input
function toggleDisplayCtxMetadataTable(show) {
    const style = show ? "display: block" : "display: none";

    const protoInput = document.getElementById("ctx-metadata-input");
    protoInput.removeAttribute("style");
    protoInput.style.cssText = style;
}
let protoMap;
let showCollection;

// IIFE to setup event listener
(function () {
    // add event listener on proto checkbox
    // if checked the uploader and proto collection view will be displayed
    // also the grpc reflection will use given proto instead of server reflection
    const protoSwitch = document.getElementById("local-proto");
    protoSwitch.addEventListener("change", function(event) {
        const { checked } = event.target;
        toggleDisplayProtoInput(checked);
    });

    // add event listener on upload proto file
    // uploaded file extension will be checked, only those with *.proto will be processed
    // add successful file to protoMap to be displayed
    const protoUploader = document.getElementById("proto-file");
    protoUploader.addEventListener("change", handleProtoUpload, false);

    // add event listener on toggle proto collection display
    // it will show / hide the proto collection
    const protoCollectionToggle = document.getElementById("proto-collection-toggle");
    protoCollectionToggle.addEventListener("click", toggleDisplayProtoCollection);

    // init map to handle proto files
    // every proto files will be unique to their name
    protoMap = new Map();

    // set proto collection display status to true
    // by default the collection will be shown
    showCollection = true;
}());

// toggle proto files display
// displaying upload button
function toggleDisplayProtoInput(show) {
    const style = show ? "display: block" : "display: none";

    const protoInput = document.getElementById("proto-input");
    protoInput.removeAttribute("style");
    protoInput.style.cssText = style;
}

// toggle proto files collection
// displaying uploaded protos collection
function toggleDisplayProtoCollection() {
    const protoCollection = document.getElementsByClassName("proto-collection")[0];
    protoCollection.removeAttribute("style");
    const protoToggle = document.getElementById("proto-collection-toggle");

    let collectionStyle = "";
    let toggleText = protoToggle.innerHTML;

    if (showCollection) {
        collectionStyle = "display: none";
        toggleText = toggleText.replace("Hide", "Show");
    } else {
        collectionStyle = "display: block";
        toggleText = toggleText.replace("Show", "Hide");
    }

    protoCollection.style.cssText = collectionStyle;
    protoToggle.innerHTML = toggleText;
    showCollection = !showCollection;
}

// handling file upload event
// add uploaded files to protoMap to avoid duplication, on file with same name the older
// file will be replaced with the latest
// if the file isn't available before, add DOM element for UI representation
// file without *.proto won't be processed
function handleProtoUpload() {
    const files = this.files;

    for (const file of files) {
        if (!file.name.endsWith(".proto")) {
            continue;
        }
        if (!protoMap.has(file.name)) {
            addProtoItem(file.name, file.size);
        }
        protoMap.set(file.name, file);
    }
}

// adding proto item to proto collection view
// give visual representation to user and access to remove unwanted uploaded files
function addProtoItem(name, size) {
    const protoItem = createProtoItem(name, size);
    const collection = document.getElementsByClassName("proto-collection")[0];
    collection.appendChild(protoItem);
}

// create dom element for proto item
// every item will be given id corresponding to it's name for easier access
function createProtoItem(name, size) {
    const item = document.createElement("div");
    item.classList.add("proto-item");
    item.id = name;

    const icon = document.createElement("img");
    icon.src = "img/file.png";
    icon.alt = "file icon";
    icon.classList.add("proto-icon");
    item.appendChild(icon);

    const desc = createProtoItemDesc(name, size);
    item.appendChild(desc);

    return item;
}

// create dom element for proto item description
function createProtoItemDesc(name, size) {
    const desc = document.createElement("div");
    desc.classList.add("proto-desc");

    const caption = document.createElement("span");
    caption.classList.add("proto-caption");
    const captionText = document.createTextNode(name.length > 15 ?
                                                name.substring(0, 12) + "..." :
                                                name);
    caption.appendChild(captionText);
    desc.appendChild(caption);

    const sizeDOM = document.createElement("span");
    sizeDOM.classList.add("proto-size");
    const sizeText = document.createTextNode(size > 1000 ?
                                             `${size/1000}kb` :
                                             `${size}b`);
    sizeDOM.appendChild(sizeText);
    desc.appendChild(sizeDOM);

    const remove = document.createElement("button");
    remove.classList.add("btn", "btn-sm", "proto-remove");
    remove.addEventListener("click", function() {removeProtoItem(name);});
    const removeIcon = document.createElement("i");
    removeIcon.classList.add("fa", "fa-trash");
    remove.appendChild(removeIcon);
    const removeText = document.createTextNode(" remove");
    remove.appendChild(removeText);
    desc.appendChild(remove);

    return desc;
}

// remove proto item based on it's ID
function removeProtoItem(name) {
    const item = document.getElementById(name);
    item.parentNode.removeChild(item);
    protoMap.delete(name);
}

// fetch all proto from protoMap
// compose it into formData to make it easier to send via ajax
function getProtos() {
    const formData = new FormData();

    for (const proto of protoMap.values()) {
        formData.append("protos", proto, proto.name);
    }

    return formData;
}

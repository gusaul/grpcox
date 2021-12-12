const objectStoreRequests = "requests";
let dbConn = null;

function gdb(){
    if (!window.indexedDB) {
        alert(`Your browser doesn't support IndexedDB`)
        return;
    }
    return  new Promise(function (success, error){
        if (dbConn !== null){
            success(dbConn)
            return
        }

        const request = indexedDB.open('grpcox',1);
        request.onerror = (event) => {
            error(request.error)
        };

        request.onsuccess = (event) => {
            // add implementation here
            console.log("success open DB")
            dbConn = event.target.result;
            success(dbConn);
        };
        // create the Contacts object store and indexes
        request.onupgradeneeded = (event) => {
            let dbConn = event.target.result;
            let store = dbConn.createObjectStore(objectStoreRequests,{
                keyPath:"name"
            });
        };
    })
}


function getRequest(name) {
    return new Promise(async function (success, error){
        let db = await gdb()
        const txn = db.transaction(objectStoreRequests, 'readwrite');
        const store = txn.objectStore(objectStoreRequests);
        let idbRequest = store.get(name);
        idbRequest.onerror = function (param) {
            error(idbRequest.error.name)
        }
        idbRequest.onsuccess = function (event) {
            success(event.target.result)
        }
    })
}

function getAllRequestKey() {
    return new Promise(async function (success, error){
        let db = await gdb();
        const txn = db.transaction(objectStoreRequests, 'readwrite');
        // get the Contacts object store
        const store = txn.objectStore(objectStoreRequests);
        let idbRequest = store.getAllKeys();
        idbRequest.onerror = function (param) {
            error(param.target)
        }
        idbRequest.onsuccess = function (event) {
            success(event.target.result)
        }
    })
}

function insertRequest(request) {
    return new Promise(async function(success, error){
        let db = await gdb()

        // create a new transaction
        const txn = db.transaction(objectStoreRequests, 'readwrite');

        const store = txn.objectStore(objectStoreRequests);

        let query = store.add(request);

        // handle success case
        query.onsuccess = function (event) {
            success('success')
        };

        // handle the error case
        query.onerror = function (event) {
            if (query.error.name === "ConstraintError") {
                error('Duplicate request name')
                return
            }
            error(query.error.name)
        }
    })
}

function updateRequest(request) {
    return new Promise(async function(success, error){
        let db = await gdb()

        // create a new transaction
        const txn = db.transaction(objectStoreRequests, 'readwrite');

        const store = txn.objectStore(objectStoreRequests);

        let query = store.put(request);

        // handle success case
        query.onsuccess = function (event) {
            success('success')
        };

        // handle the error case
        query.onerror = function (event) {
            if (query.error.name === "ConstraintError") {
                error('Duplicate request name')
                return
            }
            error(query.error.name)
        }
    })
}

function deleteRequest(name){
    return new Promise(async function (success, error){
        let db = await gdb()

        // create a new transaction
        const txn = db.transaction(objectStoreRequests, 'readwrite');

        const store = txn.objectStore(objectStoreRequests);

        let query = store.delete(name);

        // handle success case
        query.onsuccess = function (event) {
            success('success')
        };

        // handle the error case
        query.onerror = function (event) {
            error(query.error.name)
        }
    })
}
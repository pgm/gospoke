YAHOO.example.DynamicData = function() {
    var twodigits = function(x) { 
    	return (x >= 10 ? "": "0")+x;
    }

	var myFormatDateTime = function (elCell, oRecord, oColumn, oData) {
		elCell.innerHTML = oData.getFullYear() + "/" + twodigits(oData.getMonth()) + "/" + twodigits(oData.getDate()) + " " + twodigits(oData.getHours()) + ":" + twodigits(oData.getMinutes()) + ":" + twodigits(oData.getSeconds());
	}

    // Column definitions
    var myColumnDefs = [ // sortable:true enables sorting
        {key:"id", label:"ID", sortable:true},
        {key:"severity", label:"Severity", sortable:true},
        {key:"service", label:"Service", sortable:false},
        {key:"summary", label:"Summary", sortable:true},
        //{key:"timestamp", label:"Timestamp", sortable:true, formatter:myFormatDateTime}
    ];

    // Custom parser
    var stringToDate = function(sData) {
        return new Date(sData);
    };
    
    // DataSource instance
    var myDataSource = new YAHOO.util.DataSource("list-events-data?");
    myDataSource.responseType = YAHOO.util.DataSource.TYPE_JSON;
    myDataSource.responseSchema = {
        resultsList: "records",
        fields: [
            {key:"id"},
            {key:"severity"},
            {key:"service"},
            {key:"summary"},
            //{key:"timestamp", parser:stringToDate}
        ],
        metaFields: {
            totalRecords: "totalRecords" // Access to value in the server response
        }
    };
    
    var makeEventQueryUrl = function() {
    	var service_id = YAHOO.util.Dom.get('service_id').value;
    	return "service="+service_id;
    }
    
    // DataTable configuration
    var myConfigs = {
        initialRequest: makeEventQueryUrl(), // Initial request for first page of data
        dynamicData: true, // Enables dynamic server-driven data
        sortedBy : {key:"id", dir:YAHOO.widget.DataTable.CLASS_ASC}, // Sets UI initial sort arrow
        paginator: new YAHOO.widget.Paginator({ rowsPerPage:25 }) // Enables pagination 
    };
    
    // DataTable instance
    var myDataTable = new YAHOO.widget.DataTable("dynamicdata", myColumnDefs, myDataSource, myConfigs);
    // Update totalRecords on the fly with value from server
    myDataTable.handleDataReturnPayload = function(oRequest, oResponse, oPayload) {
        oPayload.totalRecords = oResponse.meta.totalRecords;
        return oPayload;
    }

    // Subscribe to events for row selection
	myDataTable.subscribe("rowMouseoverEvent", myDataTable.onEventHighlightRow);
	myDataTable.subscribe("rowMouseoutEvent", myDataTable.onEventUnhighlightRow);
	myDataTable.subscribe("rowClickEvent", myDataTable.onEventSelectRow);

    var refreshTable = function() {
				var generateRequest = myDataTable.get("generateRequest");
				var request = generateRequest(myDataTable.getState())+"&"+makeEventQueryUrl();
				
				var callback = {
					success : myDataTable.onDataReturnSetRows,
					failure : myDataTable.onDataReturnSetRows,
					scope   : myDataTable,
					argument: myDataTable.getState()
					};
				
				myDataTable.getDataSource().sendRequest(request, callback);
    }
    
    var onClearEventsButtonClick = function(event) {  
    	var rowIds = myDataTable.getSelectedRows();
    	
    	eventIds = [];
    	for(var i=0;i<rowIds.length;i++) {
    		var eventId = myDataTable.getRecord(rowIds[i]).getData("id");
    		eventIds.push("id="+encodeURIComponent(eventId));
    	}

    	var callback = {
	        success : function(o) {  
	        	refreshTable(); },
	        
    	    failure : function(o) { console.log("failure"); },
        	scope   : this,
        	argument: this
		    };

		var postData = eventIds.join("&");
    	YAHOO.util.Connect.asyncRequest('POST', 'remove-events', callback, postData);
    
    };
    
    var onRefreshEventsButtonClick = function(event) {
    	refreshTable();
    };
    
    var clearEventsButton = new YAHOO.widget.Button("clear-events", { onclick: { fn: onClearEventsButtonClick } }); 
	var refreshEventsButton = new YAHOO.widget.Button("refresh-events", { onclick: { fn: onRefreshEventsButtonClick } }); 
	YAHOO.util.Event.addListener("service_filter", "change", refreshTable);
    
}();

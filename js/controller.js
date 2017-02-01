app.controller("FrackController", ['$scope', '$location', '$routeParams', '$http', function($scope, $location, $routeParams, $http) {
    $scope.vmaWanted = [];
    $scope.allVMAS = range(12, 22);
    $scope.rootUrl = $location.absUrl().replace($location.url(), "");

    if ($routeParams.name) {
        $scope.selectedProgram = $routeParams.name;
    }

    // By default use 12-18, if we have a 12:18 in the router config parse it to
    // range int 12.13.14....18...etc. if only one number then pass it straight
    if ($routeParams.vma) {
        if ($routeParams.vma.indexOf(":") == -1) {
            $scope.selectedVMA = $routeParams.vma;
            $scope.vmaWanted = [parseInt($routeParams.vma)];
        } else {
            var sp = $routeParams.vma.split(':');
            $scope.vmaWanted = range(sp[0], sp[1]);
        }
    } else {
        $scope.vmaWanted = $scope.allVMAS;
    }

    if (!$scope.programs) {
        var res = $http.get('config/frack.yaml');
	    res.then(function succesCB(r) {
            $scope.programs = jsyaml.load(r.data);;
            $scope.programNames = [];
            for (var p of $scope.programs)
                $scope.programNames.push(p.name);
	    }, function errorCB(r) {
            console.log(r);
        });
    }

    $scope.submit = function(tourl) {
        var t = "", p ="";
        if ( $scope.selectedVMA) {
            t = $scope.selectedVMA;
        } else if ($scope.vmaWanted.length > 1) {
            t = $scope.vmaWanted[0] + ":" + $scope.vmaWanted[$scope.vmaWanted.length - 1];
        } else {
            t = $scope.allVMAS[0] + ":" + $scope.allVMAS[$scope.allVMAS.length - 1];; //default
        }

        p = tourl ? tourl : $scope.selectedProgram;
        if (typeof(p) == "undefined")
            return false;

        $location.path("/workout/" + p + "/vma/" + t);
        return true; //make emacs happy
    };

}]);

app.controller("TabsCtrl", ['$scope', function($scope) {
    $scope.tabs = [
        { link : '#!/home', label : 'Home' },
        { link : '#!/help', label : 'Comment calculer sa VMA' },
    ];

    $scope.selectedTab = $scope.tabs[0];
    $scope.setSelectedTab = function(tab) {
        $scope.selectedTab = tab;
    };

    $scope.tabClass = function(tab) {
        if ($scope.selectedTab == tab) {
            return "active";
        } else {
            return "";
        }
    };
}]);

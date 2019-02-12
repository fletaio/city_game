Math.log2 = Math.log2 || function(x){return Math.log(x)*Math.LOG2E;};
Math.log10 = Math.log10 || function(x) {return Math.log(x) * Math.LOG10E;};
if (typeof Symbol === "undefined") {
    Symbol = function () {
        function s4() {
            return Math.floor((1 + Math.random()) * 0x10000)
                .toString(16)
                .substring(1);
        }
        return s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4();
    }
}

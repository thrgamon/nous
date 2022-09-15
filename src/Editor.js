var HyperMD = require("hypermd")
// some .css files will be implicitly imported by "hypermd"
// you may get the list with HyperMD Dependency Walker

// Load these modes if you want highlighting ...
require("codemirror/mode/htmlmixed/htmlmixed") // for embedded HTML
require("codemirror/mode/stex/stex") // for Math TeX Formular
require("codemirror/mode/yaml/yaml") // for Front Matters

// Load PowerPacks if you want to utilize 3rd-party libs
require("hypermd/powerpack/fold-math-with-katex") // implicitly requires "katex"
require("hypermd/powerpack/hover-with-marked") // implicitly requires "marked"
// and other power packs...
// Power packs need 3rd-party libraries. Don't forget to install them!

var myTextarea = document.getElementById("myTextarea")
var cm = HyperMD.fromTextArea(myTextarea, {
  /* optional editor options here */
  hmdModeLoader: false, // see NOTEs below
})

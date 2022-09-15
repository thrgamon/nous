import { useState, useEffect, useRef } from "react";
import { Excalidraw } from "@excalidraw/excalidraw";

function Scali() {
  const excalidrawRef = useRef(null);
  const [viewModeEnabled, setViewModeEnabled] = useState(false);
  const [zenModeEnabled, setZenModeEnabled] = useState(false);
  const [gridModeEnabled, setGridModeEnabled] = useState(false);

  const generateTextInBox = (text) => {
    const textId = crypto.randomUUID()
    const boxId = crypto.randomUUID()
    const elements = [{
      "id": boxId,
      "type": "rectangle",
      "x": 0.3671875,
      "y": -777.7734374999999,
      "width": 227.4375,
      "height": 329.9140625,
      "angle": 0,
      "strokeColor": "#000000",
      "backgroundColor": "transparent",
      "fillStyle": "hachure",
      "strokeWidth": 1,
      "strokeStyle": "solid",
      "roughness": 1,
      "opacity": 100,
      "groupIds": [],
      "strokeSharpness": "sharp",
      "seed": 694178520,
      "version": 1,
      "versionNonce": 578055896,
      "isDeleted": false,
      "boundElements": [
        {
          "type": "text",
          "id": textId
        }
      ],
      "updated": 1663152627499,
      "link": null,
      "locked": false
    },
    {
      "id": textId,
      "type": "text",
      "x": 5.3671875,
      "y": -225.3164062499999,
      "width": 217,
      "height": 25,
      "angle": 0,
      "strokeColor": "#000000",
      "backgroundColor": "transparent",
      "fillStyle": "hachure",
      "strokeWidth": 1,
      "strokeStyle": "solid",
      "roughness": 1,
      "opacity": 100,
      "groupIds": [],
      "strokeSharpness": "sharp",
      "seed": 891965656,
      "version": 47,
      "versionNonce": 1081117352,
      "isDeleted": false,
      "boundElements": null,
      "updated": 1663152627499,
      "link": null,
      "locked": false,
      "text": text,
      "fontSize": 20,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "baseline": 18,
      "containerId": boxId,
      "originalText": text
    }]
    return elements
  }

  const initialData = {
    elements: generateTextInBox('wow foo eat dust')
  }

  return (
    <div>
        <div className="excalidraw-wrapper">
          <Excalidraw
            ref={excalidrawRef}
            initialData={initialData}
            onChange={(elements, state) =>
              console.log("Elements :", elements, "State : ", state)
            }
            onPointerUpdate={(payload) => console.log(payload)}
            onCollabButtonClick={() =>
              window.alert("You clicked on collab button")
            }
            viewModeEnabled={viewModeEnabled}
            zenModeEnabled={zenModeEnabled}
            gridModeEnabled={gridModeEnabled}
          />
        </div>
    </div>
  );
}

export default Scali;

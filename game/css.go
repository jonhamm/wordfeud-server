package game

import (
	"os"
	"path"
)

const cssStyles = `
:root {

    /* color vars */
        --tl: #4863A6; /* triple letter value */
        --tw: #7F3A3E; /* triple word value */
        --dl: #729B67; /* double letter value */
        --dw: #B77320; /* double word value */
        --ct: #64415F; /* center color */
        --letter: #F2EEE9; /* player letters background */
        --lettercolor: #0d1116 /* invert(var(--letter)); */
        --lightmono: #2C2F36;
        --darkmono: #1D1D1D;
        --boardheaderbackground: #a0a0a0;

    /* size vars */
        --tablesize: 640px;
        --cellsize: var(--tablesize) / 16;
}

.navigate  {
	font-size: 16px;
	padding: 10px 24px;
	border-radius: 8px;
	background-color: #a0a0a0;
	color: white;
}

.disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.board-h-hdr {
  background: var(--boardheaderbackground);
}

.board-v-hdr {
  background: var(--boardheaderbackground);
}

#board {
  width: var(--tablesize);
  height: var(--tablesize);
  margin: 0;
  background: var(--darkmono);
  tr {
    height: var(--cellsize);
      margin: 1em auto;
  }
  td {
    position: relative;
    height: var(--cellsize);
    width: var(--cellsize);
    padding: 0;
    border: 1px solid var(--darkmono);
    border-radius: 4px;
    border-collapse: collapse;
    text-align: center;
    text-transform: uppercase;
    font-weight: bold;
    background: var(--lightmono);
    &.dragover {
      background: lighten(var(--lightmono), 10%);
      border-color: var(--lightmono);
    }
    &.tl {
      background: var(--tl);
      &:before {
        content: "tl";
      }
      &.dragover {
        background: lighten(var(--tl), 10%);
        border-color: var(--tl);
      }
    }
    &.tw {
      background: var(--tw);
      &:before {
        content: "tw";
      }
      &.dragover {
        background: lighten(var(--tw), 10%);
        border-color: var(--tw);
      }
    }
    &.dl {
      background: var(--dl);
      &:before {
        content: "dl";
      }
      &.dragover {
        background: lighten(var(--dl), 10%);
        border-color: var(--dl);
      }
    }
    &.dw {
      background: var(--dw);
      &:before {
        content: "dw";
      }
      &.dragover {
        background: lighten(var(--dw), 10%);
        border-color: var(--dw);
      }
    }
    &.ct {
      background: var(--ct);
      &:before {
        content: "☘";
      }
      &.dragover {
        background: lighten(var(--ct), 10%);
        border-color: var(--ct);
      }
    }
  }
}
`

func writeHtmlStyles(dirName string) error {
	var err error
	var f *os.File

	fileName := "styles.css"
	tmpFileName := fileName + "~"
	filePath := path.Join(dirName, fileName)
	tmpFilePath := path.Join(dirName, tmpFileName)
	if f, err = os.Create(tmpFilePath); err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(f.Name()) // clean up
		}
	}()

	if _, err = f.WriteString(cssStyles); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return err
	}
	f = nil
	return nil
}

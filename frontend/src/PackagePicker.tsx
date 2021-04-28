import React, { useEffect, useState } from "react";
import * as R from 'ramda';
import {useCombobox} from 'downshift'

interface Props {
  packages: string[],
  onSubmit: (pkg: string) => void
}

export default function PackagePicker(props: Props) {
  const [filteredPackages, setFilteredPackages] = useState([]);
  const { packages } = props;

  const {
    isOpen,
    inputValue,
    getMenuProps,
    getInputProps,
    getComboboxProps,
    highlightedIndex,
    getItemProps,
  } = useCombobox({
    items: filteredPackages,
    onSelectedItemChange: ({selectedItem}) => {
      props.onSubmit(selectedItem);
    },
    onInputValueChange: ({inputValue}) => {
      setFilteredPackages(inputValue.length < 2 ? [] : 
        R.take(10, packages.filter(item => item.toLowerCase().includes(inputValue.toLowerCase()))))
    },
  })

  return (
    <div {...getComboboxProps()}>
      <div>Pick a package</div> 
      <form onSubmit={(e) => {
          e.preventDefault();
          props.onSubmit(inputValue)}
        }>
          <input type="text" {...getInputProps()} />
          <ul className="Autocomplete" {...getMenuProps()}>
            {isOpen &&
              filteredPackages.map((item, index) => (
                <li
                  style={
                    highlightedIndex === index ? {backgroundColor: '#bde4ff'} : {}
                  }
                  key={`${item}${index}`}
                  {...getItemProps({item, index})}
                >
                  {item}
                </li>
              ))}
          </ul>
        <button>Go!</button>
      </form>
    </div>
  )
}
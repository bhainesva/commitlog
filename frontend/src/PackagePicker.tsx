import React, { useEffect, useState } from "react";
import * as R from 'ramda';
import {useCombobox} from 'downshift'
import './scss/PackagePicker.scss';

interface Props {
  packages: string[],
  onSubmit: (pkg: string) => void
  simple?: boolean,
}

export default function PackagePicker(props: Props) {
  const [filteredPackages, setFilteredPackages] = useState([]);
  const { packages, simple, onSubmit } = props;

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
      onSubmit(selectedItem);
    },
    onInputValueChange: ({inputValue}) => {
      setFilteredPackages(inputValue.length < 2 ? [] : 
        R.take(10, packages.filter(item => item.toLowerCase().includes(inputValue.toLowerCase()))))
    },
  })

  const className = simple ? `PackagePicker PackagePicker--simple` : "PackagePicker";

  const autocompleteOptions = filteredPackages.map((item, index) => (
    <li
      className="Autocomplete-option"
      style={
        highlightedIndex === index ? {backgroundColor: '#bde4ff'} : {}
      }
      key={`${item}${index}`}
      {...getItemProps({item, index})}
    >
      {item}
    </li>
  ))

  return (
    <div className={className} {...getComboboxProps()}>
      {!simple && <div className="PackagePicker-label">Choose a package</div>}
      <form className="PackagePicker-form" onSubmit={(e) => {
          e.preventDefault();
          props.onSubmit(inputValue)}
        }>
          <input type="text" {...getInputProps()} className="PackagePicker-input" />
            {isOpen && 
              <ul className="Autocomplete" {...getMenuProps()}>
                {autocompleteOptions}
              </ul>
            }
        <button className="PackagePicker-submit">Go!</button>
      </form>
    </div>
  )
}
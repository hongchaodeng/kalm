import {
  Chip,
  createStyles,
  OutlinedTextFieldProps,
  PropTypes,
  TextField,
  TextFieldProps,
  Theme,
  withStyles
} from "@material-ui/core";
import { Autocomplete, UseAutocompleteMultipleProps, UseAutocompleteSingleProps } from "@material-ui/lab";
import { WithStyles } from "@material-ui/styles";
import clsx from "clsx";
import Immutable from "immutable";
import React from "react";
import { WrappedFieldProps } from "redux-form";
import { ID } from "../../utils";

interface Props {
  value: string;
  onChange: any;
  options: string[];
  textFieldProps?: TextFieldProps;
}

export const MaterialTableEditAutoComplete = ({ value, onChange, options, textFieldProps, ...rest }: Props) => {
  const hanldeInputChange = (event: React.ChangeEvent<{}>, value: string) => {
    if (!event) return;
    onChange(value);
  };

  const handleOnChange = (_event: React.ChangeEvent<{}>, selectOption: string | null) => {
    if (!selectOption) return;
    onChange(selectOption);
  };

  return (
    <Autocomplete
      options={options}
      disableClearable
      freeSolo
      autoComplete
      inputValue={value}
      onInputChange={hanldeInputChange}
      onChange={handleOnChange}
      renderInput={params => {
        return (
          <TextField
            {...params}
            {...textFieldProps}
            variant="outlined"
            fullWidth
            size="small"
            onKeyPress={event => {
              if (event.key === "Enter") {
                event.preventDefault();
                // TODO, submit the edit form here
              }
            }}
          />
        );
      }}
    />
  );
};

export interface ReduxFormMultiTagsFreeSoloAutoCompleteProps
  extends WrappedFieldProps,
    WithStyles<typeof styles>,
    Pick<OutlinedTextFieldProps, "placeholder"> {}

const styles = (_theme: Theme) =>
  createStyles({
    root: {}
  });

const capitalize = (s: string): string => {
  if (typeof s !== "string") return "";
  return s.charAt(0).toUpperCase() + s.slice(1);
};

const ReduxFormMultiTagsFreeSoloAutoCompleteRaw = (props: ReduxFormMultiTagsFreeSoloAutoCompleteProps) => {
  const {
    input,
    meta: { touched, invalid, error },
    classes,
    placeholder
  } = props;

  // TODO defualt hosts
  // const hosts: string[] = [""];
  const hosts: string[] = [];

  const errors = error as (string | undefined)[] | string;
  const errorsIsArray = Array.isArray(errors);

  let errorText: string | undefined = undefined;

  if (touched && invalid) {
    if (!errorsIsArray) {
      errorText = errors as string;
    } else {
      errorText = (errors as (string | undefined)[]).find(x => x !== undefined);
    }
  }

  const id = ID();

  return (
    <Autocomplete
      classes={classes}
      multiple
      autoSelect
      clearOnEscape
      freeSolo
      id={id}
      size="small"
      options={hosts}
      onFocus={input.onFocus}
      onBlur={() => {
        // https://github.com/redux-form/redux-form/issues/2768
        //
        // If a redux-form field has normilazer, the onBlur will triger normalizer.
        // This component is complex since the real values is not the input element value.
        // So if the blur event is trigger, it will set input value(wrong value) as the autocomplete value
        // As a result, Field that is using this component mush not set a normalizer.
        (input.onBlur as any)();
      }}
      // it the value is a Immutable.List, change it to an array
      value={Immutable.isCollection(input.value) ? input.value.toArray() : input.value}
      onChange={(_event: React.ChangeEvent<{}>, values) => {
        if (values) {
          input.onChange(values);
        }
      }}
      onInputChange={() => {}}
      renderTags={(value: string[], getTagProps) =>
        value.map((option: string, index: number) => {
          let color: PropTypes.Color = "default";

          if (errorsIsArray && errors[index]) {
            color = "secondary";
          }

          return <Chip variant="outlined" label={option} size="small" color={color} {...getTagProps({ index })} />;
        })
      }
      renderInput={params => {
        return (
          <TextField
            {...params}
            margin="normal"
            variant="outlined"
            error={touched && invalid}
            label={capitalize(input.name)}
            placeholder={placeholder}
            helperText={touched && invalid && errorText}
          />
        );
      }}
    />
  );
};

export const ReduxFormMultiTagsFreeSoloAutoComplete = withStyles(styles)(ReduxFormMultiTagsFreeSoloAutoCompleteRaw);

export interface KFreeSoloAutoCompleteMultiValuesProps<T>
  extends WrappedFieldProps,
    WithStyles<typeof KFreeSoloAutoCompleteMultiValuesStyles>,
    UseAutocompleteMultipleProps<T>,
    Pick<OutlinedTextFieldProps, "placeholder" | "label" | "helperText"> {}

const KFreeSoloAutoCompleteMultiValuesStyles = (theme: Theme) =>
  createStyles({
    root: {},
    error: {
      color: theme.palette.error.main,
      border: "1px solid " + theme.palette.error.main
    }
  });

// input value is Immutable.List<string>
const KFreeSoloAutoCompleteMultiValuesRaw = (props: KFreeSoloAutoCompleteMultiValuesProps<string>) => {
  const {
    input,
    label,
    options,
    helperText,
    meta: { touched, invalid, error },
    classes,
    placeholder
  } = props;

  const errors = error as (string | undefined)[] | undefined | string;
  const errorsIsArray = Array.isArray(errors);
  const errorsArray = errors as (string | undefined)[];
  let errorText: string | undefined = undefined;

  if (touched && invalid && errorsIsArray) {
    errorText = errorsArray.find(x => x !== undefined);
  }

  if (typeof errors === "string") {
    errorText = errors;
  }

  const id = ID();

  return (
    <Autocomplete
      classes={classes}
      multiple
      autoSelect
      clearOnEscape
      freeSolo
      id={id}
      size="small"
      options={options || []}
      onFocus={input.onFocus}
      onBlur={() => {
        // https://github.com/redux-form/redux-form/issues/2768
        //
        // If a redux-form field has normilazer, the onBlur will triger normalizer.
        // This component is complex since the real values is not the input element value.
        // So if the blur event is trigger, it will set input value(wrong value) as the autocomplete value
        // As a result, Field that is using this component mush not set a normalizer.
        (input.onBlur as any)();
      }}
      value={Immutable.isCollection(input.value) ? input.value.toArray() : input.value}
      onChange={(_event: React.ChangeEvent<{}>, values) => {
        if (values) {
          input.onChange(Immutable.List(values));
        }
      }}
      onInputChange={() => {}}
      renderTags={(value: string[], getTagProps) => {
        return value.map((option: string, index: number) => {
          return (
            <Chip
              variant="outlined"
              label={option}
              classes={{ root: clsx({ [classes.error]: errorsIsArray && errorsArray[index] }) }}
              size="small"
              {...getTagProps({ index })}
            />
          );
        });
      }}
      renderInput={params => {
        return (
          <TextField
            {...params}
            margin="normal"
            variant="outlined"
            error={touched && invalid}
            label={label}
            placeholder={placeholder}
            helperText={(touched && invalid && errorText) || helperText}
          />
        );
      }}
    />
  );
};

export const KFreeSoloAutoCompleteMultiValues = withStyles(KFreeSoloAutoCompleteMultiValuesStyles)(
  KFreeSoloAutoCompleteMultiValuesRaw
);

export interface KFreeSoloAutoCompleteSingleValueProps<T>
  extends WrappedFieldProps,
    WithStyles<typeof KFreeSoloAutoCompleteSingleValueStyles>,
    Pick<OutlinedTextFieldProps, "placeholder" | "label" | "helperText">,
    UseAutocompleteSingleProps<T> {}

const KFreeSoloAutoCompleteSingleValueStyles = (_theme: Theme) =>
  createStyles({
    root: {}
  });

interface KAutoCompleteOption {
  value: string;
  label: string;
  group: string;
}

function KFreeSoloAutoCompleteSingleValueRaw<T>(
  props: KFreeSoloAutoCompleteSingleValueProps<KAutoCompleteOption>
): JSX.Element {
  const {
    input,
    label,
    helperText,
    meta: { touched, invalid, error },
    classes,
    options,
    placeholder
  } = props;

  const id = ID();

  return (
    <Autocomplete
      classes={classes}
      freeSolo
      id={id}
      groupBy={option => option.group}
      // size="small"
      options={options}
      getOptionLabel={(option: any) => {
        if (option.label) {
          return option.label;
        } else {
          return option;
        }
      }}
      onFocus={input.onFocus}
      onBlur={input.onBlur}
      // onBlur={() => {
      //   // https://github.com/redux-form/redux-form/issues/2768
      //   //
      //   // If a redux-form field has normilazer, the onBlur will triger normalizer.
      //   // This component is complex since the real values is not the input element value.
      //   // So if the blur event is trigger, it will set input value(wrong value) as the autocomplete value
      //   // As a result, Field that is using this component mush not set a normalizer.
      //   (input.onBlur as any)();
      // }}
      // it the value is a Immutable.List, change it to an array
      value={input.value}
      onInputChange={(_event: any, value: string) => {
        input.onChange(value);
      }}
      // onInputChange={(...args: any[]) => {
      // console.log("onInputChange", args);
      // }}
      // onSelect={(...args: any[]) => {
      //   console.log("onSelect", args);
      //   return true;
      // }}
      // onChange={(...args: any[]) => {
      //   console.log("onChange", args);
      //   return true;
      // }}
      renderInput={params => {
        return (
          <TextField
            {...params}
            fullWidth
            variant="outlined"
            error={touched && invalid}
            label={label}
            placeholder={placeholder}
            helperText={(touched && invalid && error) || helperText}
          />
        );
      }}
    />
  );
}

export const KFreeSoloAutoCompleteSingleValue = withStyles(KFreeSoloAutoCompleteSingleValueStyles)(
  KFreeSoloAutoCompleteSingleValueRaw
);

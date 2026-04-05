import { describe, expect, it } from "vitest";
import { pathParam } from "./params.js";

describe("pathParam", () => {
  it("returns string param", () => {
    expect(pathParam("abc")).toBe("abc");
  });
  it("returns first of array", () => {
    expect(pathParam(["x", "y"])).toBe("x");
  });
  it("handles undefined", () => {
    expect(pathParam(undefined)).toBe("");
  });
});

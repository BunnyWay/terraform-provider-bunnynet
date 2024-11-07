resource "bunnynet_compute_script" "test" {
  type = "standalone"
  name = "test"

  # to import from a file, use:
  # content = file("src/script.ts")
  content = <<-CODE
    import * as BunnySDK from "https://esm.sh/@bunny.net/edgescript-sdk@0.10.0";

    BunnySDK.net.http.serve(async (request: Request): Response | Promise<Response> => {
        return new Response('<h1>Hello world!</h1>', {headers: {"Content-Type": "text/html"}});
    });
  CODE
}

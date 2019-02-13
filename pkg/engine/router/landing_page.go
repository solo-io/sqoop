package router

const landingPageTemplateString = `
<!DOCTYPE html>

<html lang="en">

<head>
    <meta charset="utf-8" />
    <title>Landing Page</title>
                <div class="links">
                    <table id="linksTable" class="ui table">
                        <thead>
                            <tr>
                                <th>
                                    <a href="">Schema</a>
                                </th>
                                <th>
                                    <a href="">Playground</a>
                                </th>
                                <th>
                                    <a href="">Query Endpoint</a>
                                </th>
                            </tr>
                        </thead>
                        <tbody>

{{range .}}
                            <tr>
                                <td style="color:white !important;">
                                    {{ .SchemaName}}
                                </td>
                                <td>
                                    <a href="{{ .RootPath }}">{{ .RootPath }}</a>
                                </td>
                                <td>
                                    <a href="{{ .QueryPath }}">{{ .QueryPath }}</a>
                                </td>
                            </tr>
{{end}}
                        </tbody>
                    </table>
                </div>
            </section>
        </article>
    </main>
</body>

</html>


`
